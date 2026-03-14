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
            $driver = app(DriverInterface::class);
            $driver->send(json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
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

        return [
            'method' => $request->method(),
            'uri' => $request->path(),
            'status' => http_response_code() ?: 200,
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
            'ip' => $request->ip(),
        ];
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

    protected static function buildGenericMeta(): array
    {
        return [
            'duration_ms' => round((microtime(true) - static::$startTime) * 1000, 2),
            'memory_mb' => round((memory_get_usage(true) - static::$startMemory) / 1024 / 1024, 2),
        ];
    }
}
