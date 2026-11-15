<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Http;

class GenerateTraffic extends Command
{
    protected $signature = 'wiretap:traffic
                            {--rounds=10 : Number of rounds to run}
                            {--delay=100 : Delay between requests in ms}';

    protected $description = 'Generate test traffic to all demo endpoints for Wiretap load testing';

    private array $endpoints = [
        ['GET', '/'],
        ['GET', '/view-demo'],
        ['GET', '/api/users'],
        ['GET', '/api/posts'],
        ['GET', '/api/posts-n-plus-one'],
        ['GET', '/api/cache-demo'],
        ['GET', '/api/transaction-demo'],
        ['GET', '/api/gate-demo'],
        ['GET', '/api/dump-demo'],
        ['GET', '/api/kitchen-sink'],
        ['GET', '/api/model-demo'],
        ['GET', '/api/http-client-demo'],
        ['GET', '/api/error'],
        ['GET', '/api/slow'],
        ['POST', '/api/orders', ['product' => 'Load Test', 'quantity' => 1]],
        ['POST', '/api/broadcast'],
        ['POST', '/api/mail-demo'],
        ['POST', '/api/notification-demo'],
        ['POST', '/api/validate-demo', ['email' => 'bad', 'name' => 'a']],
        ['POST', '/api/auth-demo'],
        ['POST', '/api/form-request-demo', ['product' => 'x', 'quantity' => 0]],
    ];

    public function handle(): int
    {
        $rounds = (int) $this->option('rounds');
        $delayMs = (int) $this->option('delay');
        $baseUrl = 'http://localhost:8000';

        $this->components->info("Generating traffic: {$rounds} rounds, {$delayMs}ms delay");
        $this->components->info(count($this->endpoints) . ' endpoints per round');

        $totalRequests = 0;
        $totalErrors = 0;
        $startTime = microtime(true);

        $commandsPerRound = 3; // inspire, test:reports, test:cleanup
        $bar = $this->output->createProgressBar($rounds * (count($this->endpoints) + $commandsPerRound));

        for ($r = 1; $r <= $rounds; $r++) {
            foreach ($this->endpoints as $endpoint) {
                $method = $endpoint[0];
                $path = $endpoint[1];
                $body = $endpoint[2] ?? [];

                try {
                    $http = Http::timeout(10)
                        ->withHeaders(['Accept' => 'application/json']);

                    if ($method === 'GET') {
                        $http->get($baseUrl . $path);
                    } else {
                        $http->post($baseUrl . $path, $body);
                    }

                    $totalRequests++;
                } catch (\Throwable $e) {
                    $totalErrors++;
                }

                $bar->advance();

                if ($delayMs > 0) {
                    usleep($delayMs * 1000);
                }
            }

            // Artisan commands (triggers command listeners)
            try {
                Artisan::call('inspire');
                $totalRequests++;
                $bar->advance();
            } catch (\Throwable) { $totalErrors++; }

            try {
                Artisan::call('test:reports');
                $totalRequests++;
                $bar->advance();
            } catch (\Throwable) { $totalErrors++; }

            try {
                Artisan::call('test:cleanup');
                $totalRequests++;
                $bar->advance();
            } catch (\Throwable) { $totalErrors++; }
        }

        $bar->finish();
        $this->newLine(2);

        $elapsed = round(microtime(true) - $startTime, 1);
        $rps = $totalRequests > 0 ? round($totalRequests / $elapsed, 1) : 0;

        $this->components->twoColumnDetail('Total requests', (string) $totalRequests);
        $this->components->twoColumnDetail('Errors', (string) $totalErrors);
        $this->components->twoColumnDetail('Duration', "{$elapsed}s");
        $this->components->twoColumnDetail('Requests/sec', (string) $rps);

        return 0;
    }
}