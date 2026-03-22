<?php

namespace Axvi\Wiretap;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class WiretapMiddleware
{
    public function handle(Request $request, Closure $next): Response
    {
        if (! config('wiretap.enabled', false)) {
            return $next($request);
        }

        if (! config('wiretap.requests.enabled', true)) {
            return $next($request);
        }

        if (Wiretap::shouldIgnore('requests', $request->path())) {
            return $next($request);
        }

        Wiretap::start();

        $data = [
            'method' => $request->method(),
            'uri' => $request->path(),
            'ip' => $request->ip(),
        ];

        $bodyFilter = config('wiretap.requests.body', 'all');
        if ($bodyFilter !== 'none') {
            $data['headers'] = collect($request->headers->all())
                ->map(fn ($v) => $v[0] ?? '')
                ->except(['cookie', 'authorization'])
                ->toArray();

            $input = $request->all();
            if (! empty($input)) {
                $filtered = Wiretap::filterData('requests', Wiretap::maskSensitive($input), 'body');
                if ($filtered !== null) {
                    $data['body'] = $filtered;
                }
            }
        }

        Wiretap::send('request.start', $data);

        return $next($request);
    }

    public function terminate(Request $request, Response $response): void
    {
        if (Wiretap::$rid === '') {
            return;
        }

        Wiretap::send('request.end', [
            'status' => $response->getStatusCode(),
            'duration_ms' => round((microtime(true) - Wiretap::$startTime) * 1000, 2),
            'memory_mb' => round(memory_get_usage(true) / 1024 / 1024, 2),
        ]);

        Wiretap::flush();
    }
}
