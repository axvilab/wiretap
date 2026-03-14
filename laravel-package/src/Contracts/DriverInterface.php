<?php

namespace Axvi\Wiretap\Contracts;

interface DriverInterface
{
    public function send(string $payload): void;
}
