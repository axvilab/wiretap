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

    'replay' => [
        'enabled' => env('WIRETAP_REPLAY_ENABLED', false),
        'secret' => env('WIRETAP_REPLAY_SECRET', ''),
    ],

    // Global field masking — applied everywhere data is captured
    'mask_fields' => [
        'password',
        'password_confirmation',
        'token',
        'secret',
        'credit_card',
    ],

    // --- Per-type configuration ---
    // 'enabled' — whether to collect this event type
    // 'ignore'  — wildcard patterns to skip (fnmatch)
    // Data filters accept: 'all', 'none', 'keys', or ['field1', 'field2']

    'requests' => [
        'enabled' => true,
        'ignore' => ['_wiretap/*', '_wiretap_test/*', '_test/*', '_debugbar/*', 'telescope/*', 'horizon/*'],
        'body' => 'all',
    ],

    'queries' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'exceptions' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'jobs' => [
        'enabled' => true,
        'ignore' => [],
        'payload' => 'all',
    ],

    'mail' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'cache' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'notifications' => [
        'enabled' => true,
        'ignore' => [],
        'data' => 'all',
    ],

    'http_client' => [
        'enabled' => true,
        'ignore' => [],
        'response_body' => 'all',
    ],

    'auth' => [
        'enabled' => true,
    ],

    'gate' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'dumps' => [
        'enabled' => true,
    ],

    'validation' => [
        'enabled' => true,
    ],

    'transactions' => [
        'enabled' => true,
    ],

    'commands' => [
        'enabled' => true,
        'ignore' => [
            'schedule:run', 'schedule:finish', 'queue:work', 'queue:listen', 'queue:restart',
            'horizon:*', 'config:*', 'route:*', 'cache:*', 'view:*',
            'package:discover', 'tinker', 'list', 'help', 'env',
        ],
    ],

    'schedule' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'logs' => [
        'enabled' => true,
        'ignore' => [],
        'context' => 'all',
    ],

    'models' => [
        'enabled' => true,
        'ignore' => [],
    ],

    'views' => [
        'enabled' => true,
        'ignore' => [],
        'data' => 'none',
        'overrides' => [
            // 'dashboard' => 'all',
            // 'welcome' => ['user', 'locale'],
            // 'emails.*' => 'none',
        ],
    ],

    'broadcasts' => [
        'enabled' => true,
        'ignore' => [],
    ],
];
