<?php

namespace Axvi\Wiretap;

use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;
use Illuminate\Support\Facades\Artisan;

class ReplayController extends Controller
{
    public function __construct()
    {
        // Verify replay secret
        $this->middleware(function (Request $request, $next) {
            $secret = config('wiretap.replay.secret', '');
            if ($secret === '') {
                return response()->json(['error' => 'replay secret not configured'], 403);
            }
            if ($request->header('X-Wiretap-Secret') !== $secret) {
                return response()->json(['error' => 'invalid secret'], 403);
            }
            return $next($request);
        });
    }

    public function retryJob(Request $request): JsonResponse
    {
        $uuid = $request->input('uuid');
        if (! $uuid) {
            return response()->json(['error' => 'uuid is required'], 400);
        }

        try {
            $exitCode = Artisan::call('queue:retry', ['id' => [$uuid]]);

            return response()->json([
                'success' => $exitCode === 0,
                'output' => trim(Artisan::output()),
            ]);
        } catch (\Throwable $e) {
            return response()->json([
                'success' => false,
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function dispatchJob(Request $request): JsonResponse
    {
        $serialized = $request->input('job_serialized');
        $overrides = $request->input('job_payload', []);

        if (! $serialized) {
            return response()->json(['error' => 'job_serialized is required'], 400);
        }

        try {
            $job = unserialize(base64_decode($serialized));

            // Apply user overrides to public properties
            foreach ($overrides as $key => $value) {
                if (property_exists($job, $key)) {
                    $reflection = new \ReflectionProperty($job, $key);
                    if ($reflection->isPublic()) {
                        $job->{$key} = $value;
                    }
                }
            }

            dispatch($job);

            return response()->json([
                'success' => true,
                'job_class' => get_class($job),
            ]);
        } catch (\Throwable $e) {
            return response()->json([
                'success' => false,
                'error' => $e->getMessage(),
            ], 500);
        }
    }

    public function replayBroadcast(Request $request): JsonResponse
    {
        $channels = $request->input('channels', []);
        $eventName = $request->input('event');
        $payload = $request->input('payload', []);

        if (! $eventName || empty($channels)) {
            return response()->json(['error' => 'event and channels are required'], 400);
        }

        try {
            $channelNames = array_map(fn ($ch) => $ch['name'] ?? $ch, $channels);
            broadcast(new ReplayBroadcastEvent($eventName, $channelNames, $payload));

            return response()->json(['success' => true, 'channels' => $channelNames]);
        } catch (\Throwable $e) {
            $error = $e->getMessage();
            if (str_contains($error, 'not defined')) {
                $error .= '. Check your broadcasting.php config — the broadcast driver connection must be defined.';
            }
            return response()->json([
                'success' => false,
                'error' => $error,
            ], 500);
        }
    }

    public function replayCommand(Request $request): JsonResponse
    {
        $command = $request->input('command');
        if (! $command) {
            return response()->json(['error' => 'command is required'], 400);
        }

        // Parse input string into arguments array
        $arguments = [];
        $input = $request->input('input', '');
        if ($input !== '') {
            // Simple parse: split by spaces, handle --key=value
            foreach (preg_split('/\s+/', $input) as $part) {
                if ($part !== '' && $part !== $command) {
                    $arguments[] = $part;
                }
            }
        }

        try {
            $exitCode = Artisan::call($command, $arguments);

            return response()->json([
                'success' => $exitCode === 0,
                'output' => trim(Artisan::output()),
                'exit_code' => $exitCode,
            ]);
        } catch (\Throwable $e) {
            return response()->json([
                'success' => false,
                'error' => $e->getMessage(),
            ], 500);
        }
    }
}