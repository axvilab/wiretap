<?php

namespace Axvi\Wiretap;

use Axvi\Wiretap\Contracts\DriverInterface;
use Axvi\Wiretap\Drivers\HttpDriver;
use Axvi\Wiretap\Drivers\RedisDriver;
use Illuminate\Cache\Events\CacheHit;
use Illuminate\Cache\Events\CacheMissed;
use Illuminate\Cache\Events\KeyForgotten;
use Illuminate\Cache\Events\KeyWritten;
use Illuminate\Contracts\Http\Kernel;
use Illuminate\Database\Events\QueryExecuted;
use Illuminate\Database\Events\TransactionBeginning;
use Illuminate\Database\Events\TransactionCommitted;
use Illuminate\Database\Events\TransactionRolledBack;
use Illuminate\Mail\Events\MessageSending;
use Illuminate\Mail\Events\MessageSent;
use Illuminate\Notifications\Events\NotificationSending;
use Illuminate\Notifications\Events\NotificationSent;
use Illuminate\Queue\Events\JobFailed;
use Illuminate\Queue\Events\JobProcessed;
use Illuminate\Queue\Events\JobProcessing;
use Illuminate\Queue\Events\JobQueued;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Event;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Queue;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\ServiceProvider;

class WiretapServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->mergeConfigFrom(__DIR__ . '/../config/wiretap.php', 'wiretap');

        $this->app->singleton(DriverInterface::class, function ($app) {
            $driver = config('wiretap.driver', 'http');

            return match ($driver) {
                'http' => new HttpDriver(
                    endpoint: config('wiretap.drivers.http.endpoint', 'http://localhost:9090/ingest'),
                    timeoutMs: config('wiretap.drivers.http.timeout_ms', 50),
                ),
                'redis' => new RedisDriver(
                    connection: config('wiretap.drivers.redis.connection', 'default'),
                    queue: config('wiretap.drivers.redis.queue', 'wiretap:events'),
                ),
                default => new HttpDriver(config('wiretap.drivers.http.endpoint')),
            };
        });
    }

    public function boot(): void
    {
        if (! config('wiretap.enabled', false)) {
            return;
        }

        $this->publishes([
            __DIR__ . '/../config/wiretap.php' => config_path('wiretap.php'),
        ], 'wiretap-config');

        if ($this->app->runningInConsole()) {
            $this->commands([GenerateReplaySecretCommand::class]);
        }

        /** @var \Illuminate\Foundation\Http\Kernel $kernel */
        $kernel = $this->app->make(Kernel::class);
        $kernel->pushMiddleware(WiretapMiddleware::class);

        $this->registerQueryListener();
        $this->registerExceptionHandler();
        $this->registerJobListeners();
        $this->registerCacheListeners();
        $this->registerMailListeners();
        $this->registerNotificationListeners();
        $this->registerHttpClientListeners();
        $this->registerTransactionListeners();
        $this->registerAuthListeners();
        $this->registerArtisanListeners();
        $this->registerScheduleListeners();
        $this->registerGateListener();
        $this->registerVarDumper();
        $this->registerValidationListener();
        $this->registerBroadcastListener();
        $this->registerLogListener();
        $this->registerReplayRoutes();
    }

    protected function registerQueryListener(): void
    {
        if (! config('wiretap.collect.queries', true)) {
            return;
        }

        DB::listen(function (QueryExecuted $query) {
            if (Wiretap::$rid === '') {
                return;
            }
            Wiretap::send('query', [
                'sql' => $query->sql,
                'bindings' => $query->bindings,
                'time_ms' => $query->time,
                'connection' => $query->connectionName,
            ]);
        });
    }

    protected function registerExceptionHandler(): void
    {
        if (! config('wiretap.collect.exceptions', true)) {
            return;
        }

        $this->app->make(\Illuminate\Contracts\Debug\ExceptionHandler::class)
            ->reportable(function (\Throwable $e) {
                if (Wiretap::$rid === '') {
                    return;
                }
                Wiretap::send('exception', [
                    'class' => get_class($e),
                    'message' => $e->getMessage(),
                    'file' => $e->getFile(),
                    'line' => $e->getLine(),
                    'trace' => array_slice($e->getTrace(), 0, 10),
                ]);
            });
    }

    protected function registerJobListeners(): void
    {
        if (! config('wiretap.collect.jobs', true)) {
            return;
        }

        Queue::before(function (JobProcessing $event) {
            if (Wiretap::shouldIgnore('jobs', $event->job->resolveName())) {
                return;
            }
            // Each job gets its own rid
            Wiretap::start();

            $jobData = [
                'job_uuid' => $event->job->uuid(),
                'job_class' => $event->job->resolveName(),
                'queue' => $event->job->getQueue(),
            ];

            // Capture job payload for replay
            if (config('wiretap.collect.job_payload', true)) {
                try {
                    $queuePayload = $event->job->payload();
                    $serialized = $queuePayload['data']['command'] ?? '';
                    if ($serialized !== '') {
                        $jobObject = unserialize($serialized);
                        $properties = Wiretap::extractJobProperties($jobObject);
                        $properties = Wiretap::maskSensitive($properties);

                        $jobData['job_payload'] = $properties;
                        $jobData['job_serialized'] = base64_encode($serialized);
                        $jobData['payload_size'] = strlen($serialized);
                    }
                } catch (\Throwable) {
                    // Can't extract payload — skip silently
                }
            }

            Wiretap::send('job.start', $jobData);
        });

        Queue::after(function (JobProcessed $event) {
            Wiretap::send('job.end', [
                'job_uuid' => $event->job->uuid(),
                'job_class' => $event->job->resolveName(),
                'status' => 'completed',
                'queue' => $event->job->getQueue(),
            ]);
            // Flush job trace as type=job
            Wiretap::flushAs('job');
        });

        Queue::failing(function (JobFailed $event) {
            Wiretap::send('job.end', [
                'job_uuid' => $event->job->uuid(),
                'job_class' => $event->job->resolveName(),
                'status' => 'failed',
                'queue' => $event->job->getQueue(),
                'error' => $event->exception->getMessage(),
            ]);
            Wiretap::flushAs('job');
        });

        Event::listen(JobQueued::class, function (JobQueued $event) {
            if (Wiretap::$rid === '') {
                return;
            }
            $job = $event->job;

            // Laravel 11+: $event->payload()['uuid'], older: $job->uuid
            $uuid = '';
            if (method_exists($event, 'payload')) {
                $payload = $event->payload();
                $uuid = $payload['uuid'] ?? '';
            }
            if ($uuid === '' && is_object($job)) {
                $uuid = $job->uuid ?? ($job->uniqueId ?? '');
            }
            if ($uuid === '' && property_exists($event, 'connectionName')) {
                // Fallback: try to get from the job's serialized data
                $uuid = data_get($job, 'uuid', '');
            }

            Wiretap::send('job.queued', [
                'job_uuid' => $uuid,
                'job_class' => is_object($job) ? get_class($job) : (string) $job,
                'queue' => is_object($job) ? ($job->queue ?? 'default') : 'default',
            ]);
        });
    }

    protected function registerCacheListeners(): void
    {
        if (! config('wiretap.collect.cache', true)) {
            return;
        }

        Event::listen(CacheHit::class, function (CacheHit $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('cache', ['key' => $event->key, 'hit' => true]);
        });

        Event::listen(CacheMissed::class, function (CacheMissed $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('cache', ['key' => $event->key, 'hit' => false]);
        });

        Event::listen(KeyWritten::class, function (KeyWritten $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('cache.write', ['key' => $event->key, 'ttl' => $event->seconds]);
        });

        Event::listen(KeyForgotten::class, function (KeyForgotten $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('cache.forget', ['key' => $event->key]);
        });
    }

    protected function registerMailListeners(): void
    {
        if (! config('wiretap.collect.mail', true)) {
            return;
        }

        Event::listen(MessageSending::class, function (MessageSending $event) {
            if (Wiretap::$rid === '') return;
            $msg = $event->message;
            Wiretap::send('mail.sending', [
                'to' => array_map(fn ($a) => $a->getAddress(), $msg->getTo()),
                'subject' => $msg->getSubject(),
            ]);
        });

        Event::listen(MessageSent::class, function (MessageSent $event) {
            if (Wiretap::$rid === '') return;
            $msg = $event->message;

            // Extract full email content for Wiretap mail viewer
            $html = '';
            $text = '';
            if ($msg->getBody() instanceof \Symfony\Component\Mime\Part\AbstractMultipartPart) {
                foreach ($msg->getBody()->getParts() as $part) {
                    if ($part instanceof \Symfony\Component\Mime\Part\TextPart) {
                        if ($part->getMediaSubtype() === 'html') {
                            $html = $part->getBody();
                        } else {
                            $text = $part->getBody();
                        }
                    }
                }
            } elseif ($msg->getBody() instanceof \Symfony\Component\Mime\Part\TextPart) {
                $body = $msg->getBody();
                if ($body->getMediaSubtype() === 'html') {
                    $html = $body->getBody();
                } else {
                    $text = $body->getBody();
                }
            }

            // Extract attachments metadata
            $attachments = [];
            if (method_exists($msg, 'getAttachments')) {
                foreach ($msg->getAttachments() as $i => $att) {
                    $attachments[] = [
                        'index' => $i,
                        'name' => $att->getFilename() ?? 'attachment',
                        'content_type' => $att->getContentType(),
                        'size' => strlen($att->getBody()),
                        'content_base64' => base64_encode($att->getBody()),
                    ];
                }
            }

            Wiretap::send('mail.sent', [
                'to' => array_map(fn ($a) => $a->getAddress(), $msg->getTo()),
                'cc' => array_map(fn ($a) => $a->getAddress(), $msg->getCc()),
                'from' => $msg->getFrom()[0]?->getAddress() ?? '',
                'subject' => $msg->getSubject(),
                'html' => $html,
                'text' => $text,
                'attachments' => $attachments,
            ]);
        });
    }

    protected function registerNotificationListeners(): void
    {
        Event::listen(NotificationSending::class, function (NotificationSending $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('notification.sending', [
                'channel' => $event->channel,
                'class' => get_class($event->notification),
                'notifiable' => get_class($event->notifiable),
            ]);
        });

        Event::listen(NotificationSent::class, function (NotificationSent $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('notification.sent', [
                'channel' => $event->channel,
                'class' => get_class($event->notification),
                'notifiable' => get_class($event->notifiable),
            ]);
        });
    }

    protected function registerHttpClientListeners(): void
    {
        try {
            Http::globalRequestMiddleware(function ($request) {
                if (Wiretap::$rid !== '') {
                    Wiretap::send('http.out.start', [
                        'method' => $request->getMethod(),
                        'url' => (string) $request->getUri(),
                    ]);
                }
                return $request;
            });

            Http::globalResponseMiddleware(function ($response) {
                if (Wiretap::$rid !== '') {
                    Wiretap::send('http.out.end', [
                        'status' => $response->getStatusCode(),
                    ]);
                }
                return $response;
            });
        } catch (\Throwable) {
            // Http client macros may not be available
        }
    }

    protected function registerTransactionListeners(): void
    {
        Event::listen(TransactionBeginning::class, function (TransactionBeginning $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('transaction', ['action' => 'begin', 'connection' => $event->connectionName]);
        });

        Event::listen(TransactionCommitted::class, function (TransactionCommitted $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('transaction', ['action' => 'commit', 'connection' => $event->connectionName]);
        });

        Event::listen(TransactionRolledBack::class, function (TransactionRolledBack $event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('transaction', ['action' => 'rollback', 'connection' => $event->connectionName]);
        });
    }

    protected function registerAuthListeners(): void
    {
        Event::listen(\Illuminate\Auth\Events\Login::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('auth.login', [
                'user_id' => $event->user->getAuthIdentifier(),
                'guard' => $event->guard ?? 'web',
            ]);
        });

        Event::listen(\Illuminate\Auth\Events\Logout::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('auth.logout', [
                'user_id' => $event->user?->getAuthIdentifier(),
                'guard' => $event->guard ?? 'web',
            ]);
        });

        Event::listen(\Illuminate\Auth\Events\Failed::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('auth.failed', [
                'guard' => $event->guard ?? 'web',
                'credentials' => array_keys($event->credentials ?? []),
            ]);
        });
    }

    protected function registerArtisanListeners(): void
    {
        Event::listen(\Illuminate\Console\Events\CommandStarting::class, function ($event) {
            if (Wiretap::shouldIgnore('commands', $event->command ?? '')) {
                return;
            }
            Wiretap::start();
            Wiretap::send('command.start', [
                'command' => $event->command,
                'input' => (string) $event->input,
            ]);
        });

        Event::listen(\Illuminate\Console\Events\CommandFinished::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('command.end', [
                'command' => $event->command,
                'exit_code' => $event->exitCode,
            ]);
            Wiretap::flushAs('command');
        });
    }

    protected function registerScheduleListeners(): void
    {
        Event::listen(\Illuminate\Console\Events\ScheduledTaskStarting::class, function ($event) {
            Wiretap::start();
            Wiretap::send('schedule.start', [
                'expression' => $event->task->expression,
                'command' => $event->task->command ?? $event->task->description,
            ]);
        });

        Event::listen(\Illuminate\Console\Events\ScheduledTaskFinished::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('schedule.end', [
                'expression' => $event->task->expression,
                'command' => $event->task->command ?? $event->task->description,
            ]);
            Wiretap::flushAs('schedule');
        });

        if (class_exists(\Illuminate\Console\Events\ScheduledTaskFailed::class)) {
            Event::listen(\Illuminate\Console\Events\ScheduledTaskFailed::class, function ($event) {
                if (Wiretap::$rid === '') return;
                Wiretap::send('schedule.failed', [
                    'expression' => $event->task->expression,
                    'error' => $event->exception->getMessage(),
                ]);
                Wiretap::flushAs('schedule');
            });
        }
    }

    protected function registerGateListener(): void
    {
        if (! class_exists(\Illuminate\Auth\Events\GateEvaluated::class)) {
            return;
        }

        Event::listen(\Illuminate\Auth\Events\GateEvaluated::class, function ($event) {
            if (Wiretap::$rid === '') return;
            Wiretap::send('gate', [
                'ability' => $event->ability,
                'result' => $event->result,
                'user_id' => $event->user?->getAuthIdentifier(),
            ]);
        });
    }

    protected function registerVarDumper(): void
    {
        if (! class_exists(\Symfony\Component\VarDumper\VarDumper::class)) {
            return;
        }

        $originalHandler = \Symfony\Component\VarDumper\VarDumper::setHandler(function ($var) use (&$originalHandler) {
            if (Wiretap::$rid !== '') {
                $trace = debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS, 5);
                $caller = $trace[1] ?? $trace[0];
                Wiretap::send('dump', [
                    'value' => print_r($var, true),
                    'file' => $caller['file'] ?? '',
                    'line' => $caller['line'] ?? 0,
                ]);
            }

            if ($originalHandler) {
                $originalHandler($var);
            }
        });
    }

    protected function registerValidationListener(): void
    {
        // Hook into exception rendering to catch ValidationException
        $this->app->make(\Illuminate\Contracts\Debug\ExceptionHandler::class)
            ->renderable(function (\Illuminate\Validation\ValidationException $e) {
                if (Wiretap::$rid !== '') {
                    Wiretap::send('validation', [
                        'errors' => $e->errors(),
                        'rules' => [],
                    ]);
                }
                return null; // let Laravel handle rendering as usual
            });
    }

    protected function registerBroadcastListener(): void
    {
        Event::listen('*', function (string $eventName, array $data) {
            if (Wiretap::$rid === '') return;

            $event = $data[0] ?? null;
            if (! is_object($event)) return;
            if (! $event instanceof \Illuminate\Contracts\Broadcasting\ShouldBroadcast) return;
            if (Wiretap::shouldIgnore('broadcasts', get_class($event))) return;

            $rawChannels = method_exists($event, 'broadcastOn')
                ? (is_array($event->broadcastOn()) ? $event->broadcastOn() : [$event->broadcastOn()])
                : [];

            $channels = [];
            foreach ($rawChannels as $ch) {
                if (is_string($ch)) {
                    $channels[] = ['name' => $ch, 'type' => 'public'];
                } elseif ($ch instanceof \Illuminate\Broadcasting\PresenceChannel) {
                    $channels[] = ['name' => $ch->name, 'type' => 'presence'];
                } elseif ($ch instanceof \Illuminate\Broadcasting\PrivateChannel) {
                    $channels[] = ['name' => $ch->name, 'type' => 'private'];
                } elseif ($ch instanceof \Illuminate\Broadcasting\Channel) {
                    $channels[] = ['name' => $ch->name, 'type' => 'public'];
                }
            }

            Wiretap::send('broadcast', [
                'event' => get_class($event),
                'channels' => $channels,
                'payload' => method_exists($event, 'broadcastWith') ? $event->broadcastWith() : [],
            ]);
        });
    }

    protected function registerLogListener(): void
    {
        if (! config('wiretap.collect.logs', true)) {
            return;
        }

        Event::listen(\Illuminate\Log\Events\MessageLogged::class, function ($event) {
            if (Wiretap::$rid === '') return;

            // Don't capture our own warnings
            if (str_starts_with($event->message, '[Wiretap]')) return;

            Wiretap::send('log', [
                'level' => $event->level,
                'message' => $event->message,
                'context' => !empty($event->context) ? array_slice($event->context, 0, 5) : [],
            ]);
        });
    }

    protected function registerReplayRoutes(): void
    {
        if (! config('wiretap.replay.enabled', false)) {
            return;
        }

        Route::prefix('_wiretap')->group(function () {
            Route::post('retry-job', [ReplayController::class, 'retryJob']);
            Route::post('dispatch-job', [ReplayController::class, 'dispatchJob']);
            Route::post('replay-broadcast', [ReplayController::class, 'replayBroadcast']);
            Route::post('replay-command', [ReplayController::class, 'replayCommand']);
        });
    }
}
