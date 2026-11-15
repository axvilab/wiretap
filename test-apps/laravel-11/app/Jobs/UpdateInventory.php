<?php

namespace App\Jobs;

use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\DB;

class UpdateInventory implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public function __construct(
        public int $orderId,
    ) {}

    public function handle(): void
    {
        // Simulate inventory queries
        DB::select('SELECT 1 as product_id, 100 as stock');
        DB::statement('SELECT 1'); // simulate UPDATE
        DB::select('SELECT 1 as product_id, 98 as stock');
    }
}
