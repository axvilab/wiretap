<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    \Illuminate\Support\Facades\Log::info('Homepage loaded', ['ip' => request()->ip()]);
    return view('welcome');
});
