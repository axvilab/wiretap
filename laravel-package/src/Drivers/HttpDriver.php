<?php

namespace Axvi\Wiretap\Drivers;

use Axvi\Wiretap\Contracts\DriverInterface;

class HttpDriver implements DriverInterface
{
    public function __construct(
        protected string $endpoint,
        protected int $timeoutMs = 50,
    ) {}

    public function send(string $payload): void
    {
        $compressed = gzencode($payload, 6);
        $headers = [
            'Content-Type: application/json',
            'Content-Encoding: gzip',
        ];

        $ingestKey = config('wiretap.ingest_key', '');
        if ($ingestKey !== '') {
            $headers[] = 'X-Wiretap-Key: ' . $ingestKey;
        }

        $ch = curl_init($this->endpoint);
        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => $compressed,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT_MS => $this->timeoutMs,
            CURLOPT_CONNECTTIMEOUT_MS => $this->timeoutMs,
            CURLOPT_NOSIGNAL => 1,
        ]);
        @curl_exec($ch);
        @curl_close($ch);
    }
}
