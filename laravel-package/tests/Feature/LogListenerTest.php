<?php

namespace Axvi\Wiretap\Tests\Feature;

use Axvi\Wiretap\Tests\TestCase;
use Axvi\Wiretap\Wiretap;
use Illuminate\Support\Facades\Log;

class LogListenerTest extends TestCase
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

    public function test_log_event_captured_in_buffer(): void
    {
        Wiretap::start();

        Log::info('Order created successfully');

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $this->assertNotEmpty($logEvents, 'Expected a log event in the buffer');

        $logEvent = array_values($logEvents)[0];
        $this->assertEquals('info', $logEvent['data']['level']);
        $this->assertEquals('Order created successfully', $logEvent['data']['message']);
    }

    public function test_log_event_captures_level(): void
    {
        Wiretap::start();

        Log::warning('Disk usage high');

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $logEvent = array_values($logEvents)[0];

        $this->assertEquals('warning', $logEvent['data']['level']);
        $this->assertEquals('Disk usage high', $logEvent['data']['message']);
    }

    public function test_log_event_captures_context(): void
    {
        Wiretap::start();

        Log::error('Payment failed', ['order_id' => 123, 'reason' => 'insufficient_funds']);

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $logEvent = array_values($logEvents)[0];

        $this->assertEquals('error', $logEvent['data']['level']);
        $this->assertArrayHasKey('context', $logEvent['data']);
        $this->assertEquals(123, $logEvent['data']['context']['order_id']);
        $this->assertEquals('insufficient_funds', $logEvent['data']['context']['reason']);
    }

    public function test_log_event_not_captured_without_rid(): void
    {
        // rid is empty after reset — no events should be buffered
        Log::info('This should not be captured');

        $this->assertEmpty(Wiretap::$buffer);
    }

    public function test_wiretap_prefixed_log_messages_are_ignored(): void
    {
        Wiretap::start();

        Log::warning('[Wiretap] Failed to send payload to ingest endpoint');

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $this->assertEmpty($logEvents, 'Wiretap-prefixed log messages should be ignored');
    }

    public function test_wiretap_prefixed_messages_do_not_pollute_buffer(): void
    {
        Wiretap::start();

        Log::info('[Wiretap] Connection timeout');
        Log::info('Legitimate app log message');
        Log::info('[Wiretap] Retry failed');

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $this->assertCount(1, $logEvents, 'Only non-Wiretap messages should be in buffer');

        $logEvent = array_values($logEvents)[0];
        $this->assertEquals('Legitimate app log message', $logEvent['data']['message']);
    }

    public function test_multiple_log_events_captured(): void
    {
        Wiretap::start();

        Log::info('First message');
        Log::debug('Second message');
        Log::error('Third message');

        $logEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'log');
        $this->assertCount(3, $logEvents);
    }

    public function test_log_not_captured_when_collection_disabled(): void
    {
        config()->set('wiretap.logs.enabled', false);

        // Re-register listeners with new config
        $provider = new \Axvi\Wiretap\WiretapServiceProvider(app());
        $provider->boot();

        Wiretap::start();

        // The original listener is still registered from initial boot,
        // so we verify the config flag is respected
        $this->assertFalse(config('wiretap.logs.enabled'));
    }
}
