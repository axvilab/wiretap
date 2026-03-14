<?php

namespace Axvi\Wiretap\Drivers;

use Axvi\Wiretap\Contracts\DriverInterface;
use Illuminate\Support\Facades\Redis;

class RedisDriver implements DriverInterface
{
    public function __construct(
        protected string $connection = 'default',
        protected string $queue = 'wiretap:events',
    ) {}

    public function send(string $payload): void
    {
        try {
            Redis::connection($this->connection)->rpush($this->queue, $payload);
        } catch (\Throwable) {
            // silently fail
        }
    }
}
