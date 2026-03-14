<?php

return [
    'enabled' => env('WIRETAP_ENABLED', false),

    'driver' => env('WIRETAP_DRIVER', 'http'),

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
        'request_body' => false,
    ],

    'mask_fields' => [
        'password',
        'password_confirmation',
        'token',
        'secret',
        'credit_card',
    ],
];
