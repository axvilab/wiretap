<?php

namespace App\Console\Commands;

use App\Jobs\ProcessOrder;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class ProcessReports extends Command
{
    protected $signature = 'test:reports';
    protected $description = 'Test command that dispatches jobs — for Wiretap tree testing';

    public function handle(): int
    {
        $this->info('Generating reports...');

        DB::select('SELECT COUNT(*) as cnt FROM users');

        $this->info('Dispatching order processing...');
        ProcessOrder::dispatch(rand(1000, 9999));

        DB::select('SELECT 1 as ok');

        $this->info('Reports done.');
        return Command::SUCCESS;
    }
}
