<?php

return [
    'enabled' => env('WIRETAP_ENABLED', false),

    'driver' => env('WIRETAP_DRIVER', 'http'),

    'ingest_key' => env('WIRETAP_INGEST_KEY', ''),

    'drivers' => [
        'http' => [
            'endpoint' => env('WIRETAP_URL', 'http://localhost:9090/ingest'),
            'timeout_ms' => env('WIRETAP_TIMEOUT_MS', 50),
        ],
        'redis' => [
            'connection' => env('WIRETAP_REDIS_CONNECTION', 'default'),
            'queue' => env('WIRETAP_REDIS_QUEUE', 'wiretap:events'),
        ],
    ],

    'collect' => [
        'queries' => true,
        'exceptions' => true,
        'jobs' => true,
        'mail' => true,
        'cache' => true,
        'notifications' => true,
        'http_client' => true,
        'auth' => true,
        'gate' => true,
        'dumps' => true,
        'validation' => true,
        'transactions' => true,
        'commands' => true,
        'schedule' => true,
        'logs' => true,
        'request_body' => true,
        'job_payload' => true,
    ],

    'replay' => [
        'enabled' => env('WIRETAP_REPLAY_ENABLED', false),
        'secret' => env('WIRETAP_REPLAY_SECRET', ''),
    ],

    'ignore' => [
        'requests' => ['_wiretap/*', '_wiretap_test/*', '_test/*', '_debugbar/*', 'telescope/*', 'horizon/*'],
        'jobs' => [],
        'commands' => [
            'schedule:run', 'schedule:finish', 'queue:work', 'queue:listen', 'queue:restart',
            'horizon:*', 'config:*', 'route:*', 'cache:*', 'view:*',
            'package:discover', 'tinker', 'list', 'help', 'env',
        ],
        'broadcasts' => [],
    ],

    'mask_fields' => [
        'password',
        'password_confirmation',
        'token',
        'secret',
        'credit_card',
    ],
];
