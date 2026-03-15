<?php

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Mail;
use Illuminate\Support\Facades\Notification;

// === Queries ===
Route::get('/users', function () {
    $users = DB::select('SELECT 1 as id, "test" as name');
    return response()->json(['users' => $users]);
});

Route::get('/posts', function () {
    $posts = DB::select('SELECT 1 as id, "Post 1" as title');
    $comments = DB::select('SELECT 1 as id, 1 as post_id, "Comment" as body');
    $tags = DB::select('SELECT 1 as id, "php" as name');
    return response()->json(['posts' => $posts, 'comments' => $comments, 'tags' => $tags]);
});

// === N+1 problem (intentional) ===
Route::get('/posts-n-plus-one', function () {
    $posts = DB::select('SELECT 1 as id, "Post" as title');
    foreach (range(1, 5) as $i) {
        DB::select('SELECT ? as id, "Author" as name', [$i]);
    }
    return response()->json(['posts' => $posts]);
});

// === Slow request ===
Route::get('/slow', function () {
    usleep(200_000);
    $result = DB::select('SELECT 1 as ok');
    return response()->json(['slow' => true, 'result' => $result]);
});

// === Exception ===
Route::get('/error', function () {
    Log::warning('About to throw test exception');
    throw new \RuntimeException('Test exception from Wiretap test-app');
});

// === Order (jobs + logs) ===
Route::post('/orders', function () {
    $orderId = rand(1000, 9999);
    Log::info('Order created', ['order_id' => $orderId]);
    DB::select('SELECT 1 as id, ? as total', [rand(10, 500)]);
    \App\Jobs\ProcessOrder::dispatch($orderId);
    Log::debug('ProcessOrder dispatched', ['order_id' => $orderId]);
    return response()->json(['order_id' => $orderId, 'status' => 'processing']);
});

// === Broadcast ===
Route::post('/broadcast', function () {
    $orderId = rand(1000, 9999);
    broadcast(new \App\Events\OrderShipped($orderId));
    return response()->json(['order_id' => $orderId, 'broadcast' => true]);
});

// === Cache (hit, miss, write, forget) ===
Route::get('/cache-demo', function () {
    Cache::put('demo_key', 'hello wiretap', 60);
    $hit = Cache::get('demo_key');
    $miss = Cache::get('nonexistent_key');
    Cache::forget('demo_key');
    return response()->json(['hit' => $hit, 'miss' => $miss]);
});

// === Transactions (begin, commit, rollback) ===
Route::get('/transaction-demo', function () {
    DB::beginTransaction();
    DB::select('SELECT 1 as inside_transaction');
    DB::commit();

    try {
        DB::beginTransaction();
        DB::select('SELECT 1 as will_rollback');
        throw new \Exception('Rollback test');
    } catch (\Exception $e) {
        DB::rollBack();
    }

    return response()->json(['transactions' => 'commit + rollback']);
});

// === Validation ===
Route::post('/validate-demo', function (\Illuminate\Http\Request $request) {
    $request->validate([
        'email' => 'required|email',
        'name' => 'required|min:3',
        'age' => 'required|integer|min:18',
    ]);
    return response()->json(['valid' => true]);
});

// === Gate ===
Route::get('/gate-demo', function () {
    Gate::define('admin', fn () => false);
    Gate::define('view-posts', fn () => true);

    $isAdmin = Gate::allows('admin');
    $canView = Gate::allows('view-posts');

    return response()->json(['admin' => $isAdmin, 'view_posts' => $canView]);
});

// === Dump (var_dump/dump/dd replacement) ===
Route::get('/dump-demo', function () {
    dump('Debug value from dump-demo route');
    dump(['user_id' => 42, 'role' => 'admin']);
    return response()->json(['dumped' => true]);
});

// === External HTTP Client ===
Route::get('/http-client-demo', function () {
    $response1 = Http::timeout(5)->get('https://httpbin.org/get', ['source' => 'wiretap']);
    $response2 = Http::timeout(5)->post('https://httpbin.org/post', ['name' => 'Wiretap', 'version' => '0.1']);
    return response()->json([
        'get_status' => $response1->status(),
        'post_status' => $response2->status(),
    ]);
});

// === Mail ===
Route::post('/mail-demo', function () {
    Mail::raw('Hello from Wiretap mail demo!', function ($msg) {
        $msg->to('user@example.com')->subject('Wiretap Mail Test');
    });
    return response()->json(['sent' => true]);
});

// === Notification ===
Route::post('/notification-demo', function () {
    Notification::route('mail', 'admin@example.com')
        ->notify(new \App\Notifications\TestNotification('Order #1234 shipped'));
    return response()->json(['notified' => true]);
});

// === Auth (login simulation) ===
Route::post('/auth-demo', function () {
    // Simulate failed login attempt
    auth()->attempt(['email' => 'test@example.com', 'password' => 'wrong']);
    return response()->json(['auth' => 'attempted']);
});

// === FormRequest validation ===
Route::post('/form-request-demo', function (\App\Http\Requests\CreateOrderRequest $request) {
    return response()->json(['valid' => true, 'data' => $request->validated()]);
});

// === All-in-one demo ===
Route::get('/kitchen-sink', function () {
    Log::info('Kitchen sink started');
    Cache::put('sink_key', 'value', 60);
    Cache::get('sink_key');
    DB::beginTransaction();
    DB::select('SELECT 1 as demo');
    DB::commit();
    Log::debug('Kitchen sink completed');
    return response()->json(['kitchen_sink' => true]);
});

// === Health ===
Route::get('/health', function () {
    return response()->json(['status' => 'ok']);
});
