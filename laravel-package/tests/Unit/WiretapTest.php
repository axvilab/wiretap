<?php

namespace Axvi\Wiretap\Tests\Unit;

use Axvi\Wiretap\Wiretap;
use Axvi\Wiretap\Tests\TestCase;

class WiretapTest extends TestCase
{
    // --- shouldIgnore ---

    public function test_should_ignore_exact_match(): void
    {
        config()->set('wiretap.ignore.commands', ['tinker', 'list']);

        $this->assertTrue(Wiretap::shouldIgnore('commands', 'tinker'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'list'));
        $this->assertFalse(Wiretap::shouldIgnore('commands', 'inspire'));
    }

    public function test_should_ignore_wildcard(): void
    {
        config()->set('wiretap.ignore.commands', ['horizon:*', 'config:*']);

        $this->assertTrue(Wiretap::shouldIgnore('commands', 'horizon:work'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'horizon:pause'));
        $this->assertTrue(Wiretap::shouldIgnore('commands', 'config:clear'));
        $this->assertFalse(Wiretap::shouldIgnore('commands', 'migrate'));
    }

    public function test_should_ignore_request_paths(): void
    {
        config()->set('wiretap.ignore.requests', ['_wiretap/*', 'health']);

        $this->assertTrue(Wiretap::shouldIgnore('requests', '_wiretap/replay-request'));
        $this->assertTrue(Wiretap::shouldIgnore('requests', '_wiretap/retry-job'));
        $this->assertTrue(Wiretap::shouldIgnore('requests', 'health'));
        $this->assertFalse(Wiretap::shouldIgnore('requests', 'api/orders'));
    }

    public function test_should_ignore_empty_patterns(): void
    {
        config()->set('wiretap.ignore.jobs', []);

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
}
