<!DOCTYPE html>
<html>
<head><title>@yield('title', 'Wiretap Test')</title></head>
<body>
    @include('components.nav')
    <main>@yield('content')</main>
    @include('components.footer')
</body>
</html>
