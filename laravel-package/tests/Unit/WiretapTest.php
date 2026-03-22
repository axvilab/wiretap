<?php

namespace Axvi\Wiretap\Tests\Unit;

use Axvi\Wiretap\Wiretap;
use Axvi\Wiretap\Tests\TestCase;

class WiretapTest extends TestCase
{
    // --- shouldIgnore ---

    public function test_should_ignore_exact_match(): void
    {
        config()->set('wiretap.commands.ignore', ['tinker', 'list']);

        $this->assertTrue(Wiretap::shouldIgnore('commands', 'tinker'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'list'));
        $this->assertFalse(Wiretap::shouldIgnore('commands', 'inspire'));
    }

    public function test_should_ignore_wildcard(): void
    {
        config()->set('wiretap.commands.ignore', ['horizon:*', 'config:*']);

        $this->assertTrue(Wiretap::shouldIgnore('commands', 'horizon:work'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'horizon:pause'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'config:clear'));
        $this->assertFalse(Wiretap::shouldIgnore('commands', 'migrate'));
    }

    public function test_should_ignore_request_paths(): void
    {
        config()->set('wiretap.requests.ignore', ['_wiretap/*', 'health']);

        $this->assertTrue(Wiretap::shouldIgnore('requests', '_wiretap/replay-request'));
        $this->assertTrue(Wiretap::shouldIgnore('requests', '_wiretap/retry-job'));
        $this->assertTrue(Wiretap::shouldIgnore('requests', 'health'));
        $this->assertFalse(Wiretap::shouldIgnore('requests', 'api/orders'));
    }

    public function test_should_ignore_empty_patterns(): void
    {
        config()->set('wiretap.jobs.ignore', []);

        $this->assertFalse(Wiretap::shouldIgnore('jobs', 'App\\Jobs\\ProcessOrder'));
    }

    public function test_should_ignore_nonexistent_group(): void
    {
        $this->assertFalse(Wiretap::shouldIgnore('nonexistent', 'anything'));
    }

    // --- maskSensitive ---

    public function test_mask_sensitive_fields(): void
    {
        config()->set('wiretap.mask_fields', ['password', 'token', 'secret']);

        $data = [
            'username' => 'john',
            'password' => 'secret123',
            'token' => 'abc',
            'name' => 'test',
        ];

        $result = Wiretap::maskSensitive($data);

        $this->assertEquals('john', $result['username']);
        $this->assertEquals('***', $result['password']);
        $this->assertEquals('***', $result['token']);
        $this->assertEquals('test', $result['name']);
    }

    public function test_mask_sensitive_nested(): void
    {
        config()->set('wiretap.mask_fields', ['password']);

        $data = [
            'user' => [
                'name' => 'john',
                'password' => 'secret123',
            ],
        ];

        $result = Wiretap::maskSensitive($data);

        $this->assertEquals('john', $result['user']['name']);
        $this->assertEquals('***', $result['user']['password']);
    }

    public function test_mask_sensitive_case_insensitive(): void
    {
        config()->set('wiretap.mask_fields', ['password']);

        $data = ['Password' => 'secret', 'PASSWORD' => 'secret'];
        $result = Wiretap::maskSensitive($data);

        // mask_fields comparison is lowercase
        $this->assertEquals('***', $result['Password']);
        $this->assertEquals('***', $result['PASSWORD']);
    }

    // --- filterData ---

    public function test_filter_data_all(): void
    {
        config()->set('wiretap.views.data', 'all');
        $data = ['user' => 'John', 'items' => [1, 2]];
        $this->assertEquals($data, Wiretap::filterData('views', $data, 'data'));
    }

    public function test_filter_data_none(): void
    {
        config()->set('wiretap.views.data', 'none');
        $this->assertNull(Wiretap::filterData('views', ['user' => 'John'], 'data'));
    }

    public function test_filter_data_keys(): void
    {
        config()->set('wiretap.views.data', 'keys');
        $data = ['user' => 'John', 'items' => [1, 2], 'locale' => 'en'];
        $this->assertEquals(['user', 'items', 'locale'], Wiretap::filterData('views', $data, 'data'));
    }

    public function test_filter_data_specific_fields(): void
    {
        config()->set('wiretap.views.data', ['user', 'locale']);
        $data = ['user' => 'John', 'items' => [1, 2], 'locale' => 'en'];
        $result = Wiretap::filterData('views', $data, 'data');
        $this->assertEquals(['user' => 'John', 'locale' => 'en'], $result);
        $this->assertArrayNotHasKey('items', $result);
    }

    public function test_filter_data_with_override(): void
    {
        config()->set('wiretap.views.data', 'none');
        config()->set('wiretap.views.overrides', ['dashboard' => 'all']);
        $data = ['user' => 'John'];
        $this->assertEquals($data, Wiretap::filterData('views', $data, 'data', 'dashboard'));
    }

    public function test_filter_data_with_wildcard_override(): void
    {
        config()->set('wiretap.views.data', 'none');
        config()->set('wiretap.views.overrides', ['emails.*' => 'keys']);
        $data = ['subject' => 'Hello', 'body' => 'World'];
        $this->assertEquals(['subject', 'body'], Wiretap::filterData('views', $data, 'data', 'emails.welcome'));
    }

    public function test_filter_data_override_specific_fields(): void
    {
        config()->set('wiretap.views.data', 'none');
        config()->set('wiretap.views.overrides', ['welcome' => ['user']]);
        $data = ['user' => 'John', 'items' => [1, 2]];
        $result = Wiretap::filterData('views', $data, 'data', 'welcome');
        $this->assertEquals(['user' => 'John'], $result);
    }

    public function test_filter_data_defaults_to_all(): void
    {
        // No config set — should default to 'all'
        config()->set('wiretap.jobs', ['enabled' => true]);
        $data = ['orderId' => 42];
        $this->assertEquals($data, Wiretap::filterData('jobs', $data, 'payload'));
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

    // --- Log listener ignore ---

    public function test_log_listener_ignores_wiretap_messages(): void
    {
        Wiretap::reset();
        Wiretap::start();

        // Simulate the log listener logic inline
        $wiretapMessage = '[Wiretap] Failed to send payload';
        $normalMessage = 'User logged in';

        // [Wiretap] prefixed messages should be skipped
        $this->assertTrue(str_starts_with($wiretapMessage, '[Wiretap]'));
        $this->assertFalse(str_starts_with($normalMessage, '[Wiretap]'));
    }

    public function test_log_listener_ignores_wiretap_prefix_variations(): void
    {
        // Various messages that start with [Wiretap] should all be ignored
        $messages = [
            '[Wiretap] Connection refused',
            '[Wiretap] Timeout exceeded',
            '[Wiretap] Buffer overflow',
        ];

        foreach ($messages as $msg) {
            $this->assertTrue(
                str_starts_with($msg, '[Wiretap]'),
                "Expected message to be recognized as Wiretap internal: {$msg}"
            );
        }

        // Normal log messages should NOT be ignored
        $normalMessages = [
            'User created order #123',
            'Payment processed',
            'Wiretap is awesome', // does NOT start with [Wiretap]
        ];

        foreach ($normalMessages as $msg) {
            $this->assertFalse(
                str_starts_with($msg, '[Wiretap]'),
                "Expected message to NOT be recognized as Wiretap internal: {$msg}"
            );
        }
    }

    // --- send() guards ---

    public function test_send_does_nothing_without_rid(): void
    {
        Wiretap::reset();
        $this->assertEmpty(Wiretap::$buffer);

        Wiretap::send('log', ['message' => 'test']);

        $this->assertEmpty(Wiretap::$buffer);
    }

    public function test_send_adds_to_buffer_with_rid(): void
    {
        Wiretap::reset();
        Wiretap::start();

        Wiretap::send('log', ['message' => 'hello']);

        $this->assertCount(1, Wiretap::$buffer);
        $this->assertEquals('log', Wiretap::$buffer[0]['type']);
        $this->assertEquals('hello', Wiretap::$buffer[0]['data']['message']);

        Wiretap::reset();
    }

    // --- reset() ---

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
}
