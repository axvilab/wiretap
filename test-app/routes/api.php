<?php

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Route;

// Simple query test
Route::get('/users', function () {
    $users = DB::select('SELECT 1 as id, "test" as name');
    return response()->json(['users' => $users]);
});

// Multiple queries test
Route::get('/posts', function () {
    $posts = DB::select('SELECT 1 as id, "Post 1" as title');
    $comments = DB::select('SELECT 1 as id, 1 as post_id, "Comment" as body');
    $tags = DB::select('SELECT 1 as id, "php" as name');
    return response()->json(['posts' => $posts, 'comments' => $comments, 'tags' => $tags]);
});

// Error test
Route::get('/error', function () {
    throw new \RuntimeException('Test exception from Wiretap test-app');
});

// Slow request test
Route::get('/slow', function () {
    usleep(200_000); // 200ms
    $result = DB::select('SELECT 1 as ok');
    return response()->json(['slow' => true, 'result' => $result]);
});

// Order — dispatches job chain: ProcessOrder → UpdateInventory + email
Route::post('/orders', function () {
    $orderId = rand(1000, 9999);
    DB::select('SELECT 1 as id, ? as total', [rand(10, 500)]);
    \App\Jobs\ProcessOrder::dispatch($orderId);
    return response()->json(['order_id' => $orderId, 'status' => 'processing']);
});

// N+1 problem (intentional)
Route::get('/posts-n-plus-one', function () {
    // Simulate N+1: 1 query for posts + N queries for each post's author
    $posts = DB::select('SELECT 1 as id, "Post" as title');
    foreach (range(1, 5) as $i) {
        DB::select('SELECT ? as id, "Author" as name', [$i]);
    }
    return response()->json(['posts' => $posts]);
});

// Broadcast test — fires event on public, private, presence channels
Route::post('/broadcast', function () {
    $orderId = rand(1000, 9999);
    broadcast(new \App\Events\OrderShipped($orderId));
    return response()->json(['order_id' => $orderId, 'broadcast' => true]);
});

// Health
Route::get('/health', function () {
    return response()->json(['status' => 'ok']);
});
