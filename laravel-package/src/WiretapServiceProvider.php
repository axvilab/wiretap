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
        $this->registerModelListener();
        $this->registerViewListener();
        $this->registerReplayRoutes();
    }

    protected function registerQueryListener(): void
    {
        if (! config('wiretap.queries.enabled', true)) {
            return;
        }

        DB::listen(function (QueryExecuted $query) {
            if (Wiretap::$rid === '') {
                return;
            }
            if (Wiretap::shouldIgnore('queries', $query->sql)) {
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
        if (! config('wiretap.exceptions.enabled', true)) {
            return;
        }

        $this->app->make(\Illuminate\Contracts\Debug\ExceptionHandler::class)
            ->reportable(function (\Throwable $e) {
                if (Wiretap::$rid === '') {
                    return;
                }
                if (Wiretap::shouldIgnore('exceptions', get_class($e))) {
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
        if (! config('wiretap.jobs.enabled', true)) {
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
            try {
                $queuePayload = $event->job->payload();
                $serialized = $queuePayload['data']['command'] ?? '';
                if ($serialized !== '') {
                    $jobObject = unserialize($serialized);
                    $properties = Wiretap::extractJobProperties($jobObject);
                    $properties = Wiretap::maskSensitive($properties);

                    $filtered = Wiretap::filterData('jobs', $properties, 'payload');
                    if ($filtered !== null) {
                        $jobData['job_payload'] = $filtered;
                    }
                    $jobData['job_serialized'] = base64_encode($serialized);
                    $jobData['payload_size'] = strlen($serialized);
                }
            } catch (\Throwable) {
                // Can't extract payload — skip silently
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
        if (! config('wiretap.cache.enabled', true)) {
            return;
        }

        Event::listen(CacheHit::class, function (CacheHit $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('cache', $event->key)) return;
            Wiretap::send('cache', ['key' => $event->key, 'hit' => true]);
        });

        Event::listen(CacheMissed::class, function (CacheMissed $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('cache', $event->key)) return;
            Wiretap::send('cache', ['key' => $event->key, 'hit' => false]);
        });

        Event::listen(KeyWritten::class, function (KeyWritten $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('cache', $event->key)) return;
            Wiretap::send('cache.write', ['key' => $event->key, 'ttl' => $event->seconds]);
        });

        Event::listen(KeyForgotten::class, function (KeyForgotten $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('cache', $event->key)) return;
            Wiretap::send('cache.forget', ['key' => $event->key]);
        });
    }

    protected function registerMailListeners(): void
    {
        if (! config('wiretap.mail.enabled', true)) {
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

            $mailable = $event->data['__laravel_notification'] ?? null;
            if ($mailable !== null && Wiretap::shouldIgnore('mail', is_object($mailable) ? get_class($mailable) : (string) $mailable)) return;

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
        if (! config('wiretap.notifications.enabled', true)) {
            return;
        }

        Event::listen(NotificationSending::class, function (NotificationSending $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('notifications', get_class($event->notification))) return;
            Wiretap::send('notification.sending', [
                'channel' => $event->channel,
                'class' => get_class($event->notification),
                'notifiable' => is_object($event->notifiable) ? get_class($event->notifiable) : (string) $event->notifiable,
            ]);
        });

        Event::listen(NotificationSent::class, function (NotificationSent $event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('notifications', get_class($event->notification))) return;

            $data = [
                'channel' => $event->channel,
                'class' => get_class($event->notification),
                'notifiable' => is_object($event->notifiable) ? get_class($event->notifiable) : (string) $event->notifiable,
            ];

            // Capture notification properties (constructor args)
            try {
                $properties = Wiretap::extractJobProperties($event->notification);
                $properties = Wiretap::maskSensitive($properties);
                $filtered = Wiretap::filterData('notifications', $properties, 'data');
                if ($filtered !== null && ! empty($filtered)) {
                    $data['data'] = $filtered;
                }
            } catch (\Throwable) {}

            Wiretap::send('notification.sent', $data);
        });
    }

    protected function registerHttpClientListeners(): void
    {
        if (! config('wiretap.http_client.enabled', true)) {
            return;
        }

        $maxResponseSize = 512 * 1024; // 512KB
        $binaryTypes = ['image/', 'video/', 'audio/', 'application/pdf', 'application/zip', 'application/gzip', 'application/octet-stream'];

        try {
            Http::globalRequestMiddleware(function ($request) {
                if (Wiretap::$rid !== '') {
                    $url = (string) $request->getUri();
                    if (Wiretap::shouldIgnore('http_client', $url)) {
                        return $request;
                    }

                    $data = [
                        'method' => $request->getMethod(),
                        'url' => $url,
                    ];

                    // Capture request headers
                    $headers = [];
                    foreach ($request->getHeaders() as $name => $values) {
                        $headers[$name] = implode(', ', $values);
                    }
                    unset($headers['Authorization'], $headers['authorization']);
                    $data['headers'] = $headers;

                    // Capture request body
                    $body = (string) $request->getBody();
                    if ($body !== '') {
                        $data['body'] = mb_substr($body, 0, 512 * 1024);
                    }

                    Wiretap::send('http.out.start', $data);
                }
                return $request;
            });

            Http::globalResponseMiddleware(function ($response) use ($maxResponseSize, $binaryTypes) {
                if (Wiretap::$rid !== '') {
                    $data = [
                        'status' => $response->getStatusCode(),
                    ];

                    // Capture response headers
                    $responseHeaders = [];
                    foreach ($response->getHeaders() as $name => $values) {
                        $responseHeaders[$name] = implode(', ', $values);
                    }
                    $data['response_headers'] = $responseHeaders;

                    // Capture response body — skip binary content
                    $responseBodyFilter = config('wiretap.http_client.response_body', 'all');
                    if ($responseBodyFilter !== 'none') {
                        $contentType = $responseHeaders['Content-Type'] ?? $responseHeaders['content-type'] ?? '';
                        $isBinary = false;
                        foreach ($binaryTypes as $type) {
                            if (str_contains(strtolower($contentType), $type)) {
                                $isBinary = true;
                                break;
                            }
                        }

                        if (! $isBinary) {
                            $body = (string) $response->getBody();
                            $size = strlen($body);
                            if ($size <= $maxResponseSize) {
                                $data['response_body'] = $body;
                            } else {
                                $data['response_body'] = mb_substr($body, 0, $maxResponseSize);
                                $data['response_truncated'] = true;
                            }
                            $data['response_size'] = $size;
                        } else {
                            $data['response_body_skipped'] = 'binary content: ' . $contentType;
                            $data['response_size'] = (int) ($responseHeaders['Content-Length'] ?? 0);
                        }
                    }

                    Wiretap::send('http.out.end', $data);
                }
                return $response;
            });
        } catch (\Throwable) {
            // Http client macros may not be available
        }
    }

    protected function registerTransactionListeners(): void
    {
        if (! config('wiretap.transactions.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.auth.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.commands.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.schedule.enabled', true)) {
            return;
        }

        Event::listen(\Illuminate\Console\Events\ScheduledTaskStarting::class, function ($event) {
            $command = $event->task->command ?? $event->task->description;
            if (Wiretap::shouldIgnore('schedule', $command)) return;
            Wiretap::start();
            Wiretap::send('schedule.start', [
                'expression' => $event->task->expression,
                'command' => $command,
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
        if (! config('wiretap.gate.enabled', true)) {
            return;
        }

        if (! class_exists(\Illuminate\Auth\Events\GateEvaluated::class)) {
            return;
        }

        Event::listen(\Illuminate\Auth\Events\GateEvaluated::class, function ($event) {
            if (Wiretap::$rid === '') return;
            if (Wiretap::shouldIgnore('gate', $event->ability)) return;
            Wiretap::send('gate', [
                'ability' => $event->ability,
                'result' => $event->result,
                'user_id' => $event->user?->getAuthIdentifier(),
            ]);
        });
    }

    protected function registerVarDumper(): void
    {
        if (! config('wiretap.dumps.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.validation.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.broadcasts.enabled', true)) {
            return;
        }

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
        if (! config('wiretap.logs.enabled', true)) {
            return;
        }

        Event::listen(\Illuminate\Log\Events\MessageLogged::class, function ($event) {
            if (Wiretap::$rid === '') return;

            // Don't capture our own warnings
            if (str_starts_with($event->message, '[Wiretap]')) return;

            if (Wiretap::shouldIgnore('logs', $event->level)) return;

            $context = !empty($event->context) ? $event->context : [];
            if (! empty($context)) {
                $filtered = Wiretap::filterData('logs', $context, 'context');
                $context = $filtered ?? [];
            }

            Wiretap::send('log', [
                'level' => $event->level,
                'message' => $event->message,
                'context' => $context,
            ]);
        });
    }

    protected function registerModelListener(): void
    {
        if (! config('wiretap.models.enabled', true)) {
            return;
        }

        $events = ['eloquent.created: *', 'eloquent.updated: *', 'eloquent.deleted: *'];
        foreach ($events as $pattern) {
            Event::listen($pattern, function (string $eventName, array $data) {
                if (Wiretap::$rid === '') return;

                $model = $data[0] ?? null;
                if (! is_object($model)) return;

                // Extract action from event name: "eloquent.created: App\Models\User" → "created"
                $action = explode('.', explode(': ', $eventName)[0])[1] ?? 'unknown';
                $class = get_class($model);
                $key = $model->getKey();

                if (Wiretap::shouldIgnore('models', $class)) return;

                Wiretap::send('model', [
                    'action' => $action,
                    'model' => $class,
                    'key' => $key,
                ]);
            });
        }
    }

    protected function registerViewListener(): void
    {
        if (! config('wiretap.views.enabled', true)) {
            return;
        }

        $this->app['view']->composer('*', function ($view) {
            if (Wiretap::$rid === '') return;

            $name = $view->getName();

            if (Wiretap::shouldIgnore('views', $name)) return;

            $eventData = [
                'name' => $name,
                'path' => $view->getPath(),
            ];

            // Capture view data based on config
            $rawData = $view->getData();
            // Remove framework internals
            unset($rawData['__env'], $rawData['app'], $rawData['errors']);
            if (! empty($rawData)) {
                $viewData = Wiretap::filterData('views', $rawData, 'data', $name);
                if ($viewData !== null) {
                    // Convert non-scalar values to strings for safe transport
                    $safe = [];
                    foreach ($viewData as $key => $value) {
                        if (is_scalar($value) || is_null($value)) {
                            $safe[$key] = $value;
                        } elseif (is_array($value)) {
                            $safe[$key] = $value;
                        } else {
                            $safe[$key] = print_r($value, true);
                        }
                    }
                    $eventData['data'] = $safe;
                }
            }

            Wiretap::send('view', $eventData);
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
