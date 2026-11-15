<?php

namespace Axvi\Wiretap;

use Axvi\Wiretap\Contracts\DriverInterface;

class Wiretap
{
    public static string $rid = '';
    public static array $buffer = [];
    public static float $startTime = 0;
    public static float $startMemory = 0;

    public static function start(): void
    {
        static::$rid = uniqid('r_', true);
        static::$buffer = [];
        static::$startTime = microtime(true);
        static::$startMemory = memory_get_usage(true);
    }

    public static function send(string $type, array $data): void
    {
        if (static::$rid === '') {
            return;
        }

        static::$buffer[] = [
            'type' => $type,
            'ts' => microtime(true) - static::$startTime,
            'data' => $data,
        ];
    }

    public static function flush(): void
    {
        static::flushAs('request');
    }

    public static function flushAs(string $type): void
    {
        if (static::$rid === '' || empty(static::$buffer)) {
            static::reset();
            return;
        }

        $meta = match ($type) {
            'request' => static::buildRequestMeta(),
            'job' => static::buildJobMeta(),
            'command' => static::buildCommandMeta(),
            'schedule' => static::buildScheduleMeta(),
            default => static::buildGenericMeta(),
        };

        $payload = [
            'v' => 1,
            'pid' => getmypid(),
            'rid' => static::$rid,
            'ts_start' => static::$startTime,
            'ts_end' => microtime(true),
            'type' => $type,
            'meta' => $meta,
            'events' => static::$buffer,
        ];

        try {
            $encoded = json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            $maxSize = config('wiretap.max_payload_bytes', 50 * 1024 * 1024);
            if (strlen($encoded) > $maxSize) {
                // Drop oversized payload silently — log if possible
                error_log("[Wiretap] payload too large (" . strlen($encoded) . " bytes), skipping");
                static::reset();
                return;
            }
            $driver = app(DriverInterface::class);
            $driver->send($encoded);
        } catch (\Throwable) {
            // silently fail — never break the app
        }

        static::reset();
    }

    public static function reset(): void
    {
        static::$rid = '';
        static::$buffer = [];
        static::$startTime = 0;
        static::$startMemory = 0;
    }

    protected static function buildRequestMeta(): array
    {
        $request = request();

        $meta = [
            'method' => $request->method(),
            'uri' => $request->path(),
            'status' => http_response_code() ?: 200,
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
            'ip' => $request->ip(),
        ];

        // Store request body in meta for replay
        $input = $request->all();
        if (! empty($input)) {
            $filtered = static::filterData('requests', static::maskSensitive($input), 'body');
            if ($filtered !== null) {
                $meta['request_body'] = json_encode($filtered);
            }
        }

        return $meta;
    }

    protected static function buildJobMeta(): array
    {
        // Extract job info from buffer events
        $jobClass = '';
        $jobUuid = '';
        $queue = '';
        $status = 'completed';

        foreach (static::$buffer as $event) {
            if ($event['type'] === 'job.start') {
                $jobClass = $event['data']['job_class'] ?? '';
                $jobUuid = $event['data']['job_uuid'] ?? '';
                $queue = $event['data']['queue'] ?? '';
            }
            if ($event['type'] === 'job.end') {
                $status = $event['data']['status'] ?? 'completed';
            }
        }

        return [
            'job_class' => $jobClass,
            'job_uuid' => $jobUuid,
            'queue' => $queue,
            'status' => $status,
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
        ];
    }

    protected static function buildCommandMeta(): array
    {
        $command = '';
        $exitCode = 0;
        foreach (static::$buffer as $event) {
            if ($event['type'] === 'command.start') {
                $command = $event['data']['command'] ?? '';
            }
            if ($event['type'] === 'command.end') {
                $exitCode = $event['data']['exit_code'] ?? 0;
            }
        }

        return [
            'command' => $command,
            'exit_code' => $exitCode,
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
        ];
    }

    protected static function buildScheduleMeta(): array
    {
        $expression = '';
        $command = '';
        foreach (static::$buffer as $event) {
            if ($event['type'] === 'schedule.start') {
                $expression = $event['data']['expression'] ?? '';
                $command = $event['data']['command'] ?? '';
            }
        }

        return [
            'expression' => $expression,
            'command' => $command,
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
        ];
    }

    public static function shouldIgnore(string $group, string $value): bool
    {
        $patterns = config("wiretap.{$group}.ignore", []);
        return static::matchesIgnorePatterns($value, $patterns);
    }

    /**
     * Check if a value matches any of the given ignore patterns.
     * Pure function — no Laravel dependencies.
     */
    public static function matchesIgnorePatterns(string $value, array $patterns): bool
    {
        foreach ($patterns as $pattern) {
            if ($pattern === $value) {
                return true;
            }
            if (str_contains($pattern, '*') && fnmatch($pattern, $value)) {
                return true;
            }
        }
        return false;
    }

    /**
     * Filter data based on per-type config.
     *
     * @param  string      $type       Config section ('views', 'requests', 'jobs', etc.)
     * @param  array       $data       The raw data to filter
     * @param  string      $subKey     Config key within the section ('data', 'body', 'payload', 'context')
     * @param  string|null $identifier Optional identifier for per-item overrides (e.g. view name)
     * @return array|null  Filtered data, array of keys, or null ('none')
     */
    public static function filterData(string $type, array $data, string $subKey, ?string $identifier = null): ?array
    {
        $filter = config("wiretap.{$type}.{$subKey}", 'all');
        $overrides = config("wiretap.{$type}.overrides", []);

        return static::applyDataFilter($data, $filter, $overrides, $identifier);
    }

    /**
     * Apply a data filter to an array.
     * Pure function — no Laravel dependencies.
     *
     * @param  array       $data       The raw data to filter
     * @param  mixed       $filter     'all', 'none', 'keys', or ['field1', 'field2']
     * @param  array       $overrides  Per-identifier overrides (e.g. ['dashboard' => 'all'])
     * @param  string|null $identifier Optional identifier for override matching
     * @return array|null  Filtered data, array of keys, or null ('none')
     */
    public static function applyDataFilter(array $data, mixed $filter, array $overrides = [], ?string $identifier = null): ?array
    {
        if ($identifier !== null) {
            foreach ($overrides as $pattern => $override) {
                if ($pattern === $identifier || (str_contains($pattern, '*') && fnmatch($pattern, $identifier))) {
                    $filter = $override;
                    break;
                }
            }
        }

        if ($filter === 'none') {
            return null;
        }

        if ($filter === 'keys') {
            return array_keys($data);
        }

        if (is_array($filter)) {
            return array_intersect_key($data, array_flip($filter));
        }

        return $data;
    }

    public static function extractJobProperties(object $job): array
    {
        $reflection = new \ReflectionClass($job);
        $properties = [];

        foreach ($reflection->getProperties(\ReflectionProperty::IS_PUBLIC) as $prop) {
            if ($prop->isInitialized($job)) {
                $value = $prop->getValue($job);
                $properties[$prop->getName()] = is_scalar($value) || is_null($value) || is_array($value)
                    ? $value
                    : (string) $value;
            }
        }

        return $properties;
    }

    public static function maskSensitive(array $data): array
    {
        return static::maskFields($data, config('wiretap.mask_fields', []));
    }

    /**
     * Mask sensitive fields in data recursively.
     * Pure function — no Laravel dependencies.
     */
    public static function maskFields(array $data, array $fields): array
    {
        foreach ($data as $key => &$value) {
            if (is_array($value)) {
                $value = static::maskFields($value, $fields);
            } elseif (in_array(strtolower($key), $fields, true)) {
                $value = '***';
            }
        }
        return $data;
    }

    protected static function buildGenericMeta(): array
    {
        return [
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
        ];
    }
}
