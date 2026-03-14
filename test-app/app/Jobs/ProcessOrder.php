<?php

namespace App\Jobs;

use App\Mail\OrderConfirmation;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Mail;

class ProcessOrder implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public function __construct(
        public int $orderId,
    ) {}

    public function handle(): void
    {
        // Simulate queries
        DB::select('SELECT 1 as id, ? as order_id', [$this->orderId]);
        DB::select('SELECT 1 as id, "Item" as name, 10.00 as price');

        // Send HTML confirmation email
        $items = [
            ['name' => 'Wireless Headphones', 'qty' => 1, 'price' => '79.99'],
            ['name' => 'USB-C Cable (2m)', 'qty' => 2, 'price' => '12.99'],
            ['name' => 'Phone Case', 'qty' => 1, 'price' => '24.99'],
        ];
        $total = array_sum(array_map(fn($i) => (float) $i['price'] * $i['qty'], $items));

        Mail::to('customer@example.com')
            ->cc('support@example.com')
            ->send(new OrderConfirmation($this->orderId, $total, $items));

        // Dispatch child job
        UpdateInventory::dispatch($this->orderId);
    }
}
