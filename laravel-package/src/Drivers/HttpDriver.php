<?php

namespace Axvi\Wiretap\Drivers;

use Axvi\Wiretap\Contracts\DriverInterface;
use Illuminate\Support\Facades\Log;

class HttpDriver implements DriverInterface
{
    private static bool $connectionWarned = false;

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

        $result = @curl_exec($ch);
        $errno = curl_errno($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        @curl_close($ch);

        if ($errno !== 0 || ($httpCode > 0 && $httpCode >= 400)) {
            if (! self::$connectionWarned) {
                self::$connectionWarned = true;
                $error = $errno !== 0 ? curl_strerror($errno) : "HTTP {$httpCode}";
                Log::warning("[Wiretap] Cannot reach server at {$this->endpoint}: {$error}");
            }
        } else {
            self::$connectionWarned = false;
        }
    }
}
