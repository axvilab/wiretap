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

        Wiretap::start();
        Wiretap::send('request.start', [
            'method' => $request->method(),
            'uri' => $request->path(),
            'ip' => $request->ip(),
        ]);

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
