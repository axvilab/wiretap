<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    \Illuminate\Support\Facades\Log::info('Homepage loaded', ['ip' => request()->ip()]);
    return view('welcome');
});

Route::get('/view-demo', function () {
    return view('pages.dashboard', [
        'user' => 'John',
        'items' => ['Order #1', 'Order #2', 'Order #3'],
    ]);
});
