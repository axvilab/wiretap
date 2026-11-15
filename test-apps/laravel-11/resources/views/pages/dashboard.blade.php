@extends('layouts.app')

@section('title', 'Dashboard')

@section('content')
    <h1>Dashboard</h1>
    @include('components.nav')
    <p>Welcome, {{ $user ?? 'Guest' }}</p>
    @foreach($items ?? [] as $item)
        <div>{{ $item }}</div>
    @endforeach
@endsection
