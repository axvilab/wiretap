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
            $compressed = gzencode($payload, 6);

            $redis = Redis::connection($this->connection);

            // Get underlying client and push without Laravel prefix
            $client = $redis->client();
            if ($client instanceof \Redis) {
                // PhpRedis
                $prefix = $client->getOption(\Redis::OPT_PREFIX) ?: '';
                $client->setOption(\Redis::OPT_PREFIX, '');
                $client->rpush($this->queue, $compressed);
                $client->setOption(\Redis::OPT_PREFIX, $prefix);
            } elseif ($client instanceof \Predis\Client) {
                // Predis — no prefix issue, direct call
                $client->rpush($this->queue, [$compressed]);
            } else {
                // Fallback — use Laravel facade (may have prefix)
                $redis->rpush($this->queue, $compressed);
            }
        } catch (\Throwable) {
            // silently fail
        }
    }
}