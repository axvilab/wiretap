<?php

namespace Axvi\Wiretap\Tests\Unit;

use Axvi\Wiretap\Tests\TestCase;
use Axvi\Wiretap\Wiretap;

class WiretapTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();
        Wiretap::reset();
    }

    protected function tearDown(): void
    {
        Wiretap::reset();
        parent::tearDown();
    }

    // --- matchesIgnorePatterns ---

    public function test_exact_match(): void
    {
        $this->assertTrue(Wiretap::matchesIgnorePatterns('tinker', ['tinker', 'list']));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('list', ['tinker', 'list']));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('inspire', ['tinker', 'list']));
    }

    public function test_wildcard_match(): void
    {
        $patterns = ['horizon:*', 'config:*'];

        $this->assertTrue(Wiretap::matchesIgnorePatterns('horizon:work', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('horizon:pause', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('config:clear', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('migrate', $patterns));
    }

    public function test_request_path_patterns(): void
    {
        $patterns = ['_wiretap/*', 'health'];

        $this->assertTrue(Wiretap::matchesIgnorePatterns('_wiretap/replay-request', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('_wiretap/retry-job', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('health', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('api/orders', $patterns));
    }

    public function test_empty_patterns(): void
    {
        $this->assertFalse(Wiretap::matchesIgnorePatterns('App\\Jobs\\ProcessOrder', []));
    }

    public function test_query_patterns(): void
    {
        $patterns = ['select * from sessions*'];
        $this->assertTrue(Wiretap::matchesIgnorePatterns('select * from sessions where id = ?', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('select * from users', $patterns));
    }

    public function test_exception_class_patterns(): void
    {
        $patterns = ['*NotFoundException*'];
        $this->assertTrue(Wiretap::matchesIgnorePatterns('App\Exceptions\NotFoundException', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('App\Exceptions\PaymentException', $patterns));
    }

    public function test_cache_key_patterns(): void
    {
        $patterns = ['telescope:*', 'horizon:*'];
        $this->assertTrue(Wiretap::matchesIgnorePatterns('telescope:dump-123', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('horizon:stats', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('user:profile:42', $patterns));
    }

    public function test_url_patterns(): void
    {
        $patterns = ['*stripe.com*', '*analytics*'];
        $this->assertTrue(Wiretap::matchesIgnorePatterns('https://api.stripe.com/v1/charges', $patterns));
        $this->assertTrue(Wiretap::matchesIgnorePatterns('https://analytics.example.com/track', $patterns));
        $this->assertFalse(Wiretap::matchesIgnorePatterns('https://api.example.com/orders', $patterns));
    }

    // --- maskFields ---

    public function test_mask_fields(): void
    {
        $data = [
            'username' => 'john',
            'password' => 'secret123',
            'token' => 'abc',
            'name' => 'test',
        ];

        $result = Wiretap::maskFields($data, ['password', 'token', 'secret']);

        $this->assertEquals('john', $result['username']);
        $this->assertEquals('***', $result['password']);
        $this->assertEquals('***', $result['token']);
        $this->assertEquals('test', $result['name']);
    }

    public function test_mask_fields_nested(): void
    {
        $data = [
            'user' => [
                'name' => 'john',
                'password' => 'secret123',
            ],
        ];

        $result = Wiretap::maskFields($data, ['password']);

        $this->assertEquals('john', $result['user']['name']);
        $this->assertEquals('***', $result['user']['password']);
    }

    public function test_mask_fields_case_insensitive(): void
    {
        $data = ['Password' => 'secret', 'PASSWORD' => 'secret'];
        $result = Wiretap::maskFields($data, ['password']);

        $this->assertEquals('***', $result['Password']);
        $this->assertEquals('***', $result['PASSWORD']);
    }

    public function test_mask_fields_empty_list(): void
    {
        $data = ['password' => 'secret', 'name' => 'john'];
        $result = Wiretap::maskFields($data, []);

        $this->assertEquals('secret', $result['password']);
        $this->assertEquals('john', $result['name']);
    }

    // --- applyDataFilter ---

    public function test_filter_all(): void
    {
        $data = ['user' => 'John', 'items' => [1, 2]];
        $this->assertEquals($data, Wiretap::applyDataFilter($data, 'all'));
    }

    public function test_filter_none(): void
    {
        $this->assertNull(Wiretap::applyDataFilter(['user' => 'John'], 'none'));
    }

    public function test_filter_keys(): void
    {
        $data = ['user' => 'John', 'items' => [1, 2], 'locale' => 'en'];
        $this->assertEquals(['user', 'items', 'locale'], Wiretap::applyDataFilter($data, 'keys'));
    }

    public function test_filter_specific_fields(): void
    {
        $data = ['user' => 'John', 'items' => [1, 2], 'locale' => 'en'];
        $result = Wiretap::applyDataFilter($data, ['user', 'locale']);
        $this->assertEquals(['user' => 'John', 'locale' => 'en'], $result);
        $this->assertArrayNotHasKey('items', $result);
    }

    public function test_filter_with_override(): void
    {
        $data = ['user' => 'John'];
        $result = Wiretap::applyDataFilter($data, 'none', ['dashboard' => 'all'], 'dashboard');
        $this->assertEquals($data, $result);
    }

    public function test_filter_with_wildcard_override(): void
    {
        $data = ['subject' => 'Hello', 'body' => 'World'];
        $result = Wiretap::applyDataFilter($data, 'none', ['emails.*' => 'keys'], 'emails.welcome');
        $this->assertEquals(['subject', 'body'], $result);
    }

    public function test_filter_override_specific_fields(): void
    {
        $data = ['user' => 'John', 'items' => [1, 2]];
        $result = Wiretap::applyDataFilter($data, 'none', ['welcome' => ['user']], 'welcome');
        $this->assertEquals(['user' => 'John'], $result);
    }

    public function test_filter_no_override_match(): void
    {
        $data = ['user' => 'John'];
        $result = Wiretap::applyDataFilter($data, 'none', ['dashboard' => 'all'], 'welcome');
        $this->assertNull($result);
    }

    public function test_filter_defaults_to_all(): void
    {
        $data = ['orderId' => 42];
        $this->assertEquals($data, Wiretap::applyDataFilter($data, 'all'));
    }

    public function test_filter_combined_overrides(): void
    {
        $data = ['user' => 'John', 'role' => 'admin', 'items' => [1]];
        $overrides = [
            'dashboard' => 'all',
            'emails.*' => 'keys',
            'admin.users' => ['role'],
        ];

        $this->assertEquals($data, Wiretap::applyDataFilter($data, 'none', $overrides, 'dashboard'));
        $this->assertEquals(['user', 'role', 'items'], Wiretap::applyDataFilter($data, 'none', $overrides, 'emails.welcome'));
        $this->assertEquals(['role' => 'admin'], Wiretap::applyDataFilter($data, 'none', $overrides, 'admin.users'));
        $this->assertNull(Wiretap::applyDataFilter($data, 'none', $overrides, 'welcome'));
    }

    // --- extractJobProperties ---

    public function test_extract_job_properties(): void
    {
        $job = new class(42, 'shipped') {
            public int $orderId;
            public string $status;
            protected string $hidden = 'secret';
            private string $internal = 'data';

            public function __construct(int $orderId, string $status)
            {
                $this->orderId = $orderId;
                $this->status = $status;
            }
        };

        $props = Wiretap::extractJobProperties($job);

        $this->assertEquals(42, $props['orderId']);
        $this->assertEquals('shipped', $props['status']);
        $this->assertArrayNotHasKey('hidden', $props);
        $this->assertArrayNotHasKey('internal', $props);
    }

    public function test_extract_job_properties_with_array(): void
    {
        $job = new class {
            public array $ids = [1, 2, 3];
            public ?string $name = null;
        };

        $props = Wiretap::extractJobProperties($job);

        $this->assertEquals([1, 2, 3], $props['ids']);
        $this->assertNull($props['name']);
    }

    // --- send() / start() / reset() ---

    public function test_send_does_nothing_without_rid(): void
    {
        $this->assertEmpty(Wiretap::$buffer);
        Wiretap::send('log', ['message' => 'test']);
        $this->assertEmpty(Wiretap::$buffer);
    }

    public function test_send_adds_to_buffer_with_rid(): void
    {
        Wiretap::start();
        Wiretap::send('log', ['message' => 'hello']);

        $this->assertCount(1, Wiretap::$buffer);
        $this->assertEquals('log', Wiretap::$buffer[0]['type']);
        $this->assertEquals('hello', Wiretap::$buffer[0]['data']['message']);
    }

    public function test_start_generates_rid(): void
    {
        Wiretap::start();
        $this->assertNotEmpty(Wiretap::$rid);
        $this->assertStringStartsWith('r_', Wiretap::$rid);
    }

    public function test_reset_clears_state(): void
    {
        Wiretap::start();
        Wiretap::send('query', ['sql' => 'SELECT 1']);

        $this->assertNotEmpty(Wiretap::$rid);
        $this->assertNotEmpty(Wiretap::$buffer);

        Wiretap::reset();

        $this->assertEmpty(Wiretap::$rid);
        $this->assertEmpty(Wiretap::$buffer);
        $this->assertEquals(0, Wiretap::$startTime);
        $this->assertEquals(0, Wiretap::$startMemory);
    }

    public function test_send_records_timestamp(): void
    {
        Wiretap::start();
        Wiretap::send('query', ['sql' => 'SELECT 1']);

        $this->assertArrayHasKey('ts', Wiretap::$buffer[0]);
        $this->assertIsFloat(Wiretap::$buffer[0]['ts']);
        $this->assertGreaterThanOrEqual(0, Wiretap::$buffer[0]['ts']);
    }

    public function test_multiple_sends_accumulate(): void
    {
        Wiretap::start();
        Wiretap::send('query', ['sql' => 'SELECT 1']);
        Wiretap::send('query', ['sql' => 'SELECT 2']);
        Wiretap::send('log', ['message' => 'hello']);

        $this->assertCount(3, Wiretap::$buffer);
    }

    // --- wiretap prefix detection ---

    public function test_wiretap_prefix_detected(): void
    {
        $this->assertTrue(str_starts_with('[Wiretap] Connection refused', '[Wiretap]'));
        $this->assertTrue(str_starts_with('[Wiretap] Timeout exceeded', '[Wiretap]'));
        $this->assertFalse(str_starts_with('User created order #123', '[Wiretap]'));
        $this->assertFalse(str_starts_with('Wiretap is awesome', '[Wiretap]'));
    }
}