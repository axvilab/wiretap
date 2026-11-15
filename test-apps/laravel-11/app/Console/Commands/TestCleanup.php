<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class TestCleanup extends Command
{
    protected $signature = 'test:cleanup';
    protected $description = 'Test command for Wiretap — simulates cleanup task';

    public function handle(): int
    {
        $this->info('Running cleanup...');

        DB::select('SELECT COUNT(*) as cnt FROM users');
        DB::select('SELECT 1 as ok');

        usleep(50_000); // 50ms

        $this->info('Cleanup complete.');

        return Command::SUCCESS;
    }
}
