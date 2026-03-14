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

// Health
Route::get('/health', function () {
    return response()->json(['status' => 'ok']);
});
