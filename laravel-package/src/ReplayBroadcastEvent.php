<?php

namespace Axvi\Wiretap;

use Illuminate\Broadcasting\Channel;
use Illuminate\Contracts\Broadcasting\ShouldBroadcastNow;

class ReplayBroadcastEvent implements ShouldBroadcastNow
{
    public function __construct(
        private string $eventName,
        private array $channelNames,
        private array $payload,
    ) {}

    public function broadcastOn(): array
    {
        return array_map(fn ($name) => new Channel($name), $this->channelNames);
    }

    public function broadcastAs(): string
    {
        return $this->eventName;
    }

    public function broadcastWith(): array
    {
        return $this->payload;
    }
}