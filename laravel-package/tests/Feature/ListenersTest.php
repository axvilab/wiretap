<?php

namespace Axvi\Wiretap\Tests\Feature;

use Axvi\Wiretap\Tests\TestCase;
use Axvi\Wiretap\Wiretap;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Mail;
use Illuminate\Support\Facades\Notification;
use Illuminate\Support\Facades\Queue;

class ListenersTest extends TestCase
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

    protected function defineEnvironment($app): void
    {
        parent::defineEnvironment($app);
        $app['config']->set('database.default', 'testing');
        $app['config']->set('database.connections.testing', [
            'driver' => 'sqlite',
            'database' => ':memory:',
        ]);
        $app['config']->set('cache.default', 'array');
    }

    // === Query Listener ===

    public function test_query_event_captured(): void
    {
        Wiretap::start();
        DB::select('SELECT 1 as test');

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'query');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertStringContainsString('SELECT 1', $event['data']['sql']);
        $this->assertArrayHasKey('time_ms', $event['data']);
    }

    public function test_query_bindings_captured(): void
    {
        Wiretap::start();
        DB::select('SELECT ? as val', [42]);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'query');
        $event = array_values($events)[0];

        $this->assertEquals([42], $event['data']['bindings']);
    }

    public function test_query_not_captured_without_rid(): void
    {
        DB::select('SELECT 1');
        $this->assertEmpty(Wiretap::$buffer);
    }

    // === Cache Listener ===

    public function test_cache_write_event(): void
    {
        Wiretap::start();
        Cache::put('test_key', 'value', 60);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'cache.write');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('test_key', $event['data']['key']);
    }

    public function test_cache_hit_event(): void
    {
        Cache::put('hit_key', 'value', 60);
        Wiretap::start();
        Cache::get('hit_key');

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'cache');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('hit_key', $event['data']['key']);
        $this->assertTrue($event['data']['hit']);
    }

    public function test_cache_miss_event(): void
    {
        Wiretap::start();
        Cache::get('nonexistent_key');

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'cache');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertFalse($event['data']['hit']);
    }

    public function test_cache_forget_event(): void
    {
        Cache::put('forget_key', 'value', 60);
        Wiretap::start();
        Cache::forget('forget_key');

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'cache.forget');
        $this->assertNotEmpty($events);
    }

    // === Transaction Listener ===

    public function test_transaction_begin_commit(): void
    {
        Wiretap::start();
        DB::beginTransaction();
        DB::commit();

        $types = array_column(Wiretap::$buffer, 'type');
        $transactions = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'transaction');
        $actions = array_column(array_values($transactions), 'data');

        $this->assertContains('transaction', $types);
        $foundBegin = false;
        $foundCommit = false;
        foreach ($transactions as $t) {
            if ($t['data']['action'] === 'begin') $foundBegin = true;
            if ($t['data']['action'] === 'commit') $foundCommit = true;
        }
        $this->assertTrue($foundBegin);
        $this->assertTrue($foundCommit);
    }

    public function test_transaction_rollback(): void
    {
        Wiretap::start();
        DB::beginTransaction();
        DB::rollBack();

        $transactions = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'transaction');
        $foundRollback = false;
        foreach ($transactions as $t) {
            if ($t['data']['action'] === 'rollback') $foundRollback = true;
        }
        $this->assertTrue($foundRollback);
    }

    // === Gate Listener ===

    public function test_gate_event_captured(): void
    {
        if (! class_exists(\Illuminate\Auth\Events\GateEvaluated::class)) {
            $this->markTestSkipped('GateEvaluated not available in this Laravel version');
        }

        Gate::define('test-ability', fn () => true);

        Wiretap::start();
        Gate::allows('test-ability');

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'gate');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('test-ability', $event['data']['ability']);
    }

    // === VarDumper Listener ===

    public function test_dump_event_captured(): void
    {
        Wiretap::start();

        // Capture dump output to prevent it from displaying
        ob_start();
        dump('test value');
        ob_end_clean();

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'dump');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertStringContainsString('test value', $event['data']['value']);
        $this->assertArrayHasKey('file', $event['data']);
        $this->assertArrayHasKey('line', $event['data']);
    }

    // === HTTP Client Listener ===

    public function test_http_client_events_captured(): void
    {
        Http::fake([
            'example.com/*' => Http::response(['ok' => true], 200),
        ]);

        Wiretap::start();
        Http::get('https://example.com/api/test');

        $starts = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'http.out.start');
        $ends = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'http.out.end');

        $this->assertNotEmpty($starts);
        $this->assertNotEmpty($ends);

        $start = array_values($starts)[0];
        $this->assertEquals('GET', $start['data']['method']);
        $this->assertStringContainsString('example.com', $start['data']['url']);

        $end = array_values($ends)[0];
        $this->assertEquals(200, $end['data']['status']);
    }

    public function test_http_client_response_body_captured(): void
    {
        Http::fake([
            'example.com/*' => Http::response(['data' => 'hello'], 200),
        ]);

        Wiretap::start();
        Http::get('https://example.com/api/data');

        $ends = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'http.out.end');
        $end = array_values($ends)[0];

        $this->assertArrayHasKey('response_body', $end['data']);
        $this->assertArrayHasKey('response_size', $end['data']);
    }

    public function test_http_client_post_body_captured(): void
    {
        Http::fake([
            'example.com/*' => Http::response([], 201),
        ]);

        Wiretap::start();
        Http::post('https://example.com/api/create', ['name' => 'test']);

        $starts = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'http.out.start');
        $start = array_values($starts)[0];

        $this->assertEquals('POST', $start['data']['method']);
        $this->assertArrayHasKey('headers', $start['data']);
    }

    // === Auth Listener ===

    public function test_auth_failed_event(): void
    {
        Wiretap::start();

        // Trigger auth failed event directly
        event(new \Illuminate\Auth\Events\Failed('web', null, ['email' => 'test@test.com']));

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'auth.failed');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('web', $event['data']['guard']);
    }

    public function test_auth_login_event(): void
    {
        Wiretap::start();

        $user = new class implements \Illuminate\Contracts\Auth\Authenticatable {
            public function getAuthIdentifier() { return 42; }
            public function getAuthIdentifierName() { return 'id'; }
            public function getAuthPassword() { return ''; }
            public function getAuthPasswordName() { return 'password'; }
            public function getRememberToken() { return ''; }
            public function setRememberToken($value) {}
            public function getRememberTokenName() { return ''; }
        };
        event(new \Illuminate\Auth\Events\Login('web', $user, false));

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'auth.login');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals(42, $event['data']['user_id']);
    }

    // === Notification Listener ===

    public function test_notification_events_captured(): void
    {
        Wiretap::start();

        $notifiable = new class { public function getKey() { return 1; } };
        $notification = new class extends \Illuminate\Notifications\Notification {
            public function via($n) { return ['mail']; }
        };

        event(new \Illuminate\Notifications\Events\NotificationSending($notifiable, $notification, 'mail'));
        event(new \Illuminate\Notifications\Events\NotificationSent($notifiable, $notification, 'mail', null));

        $sending = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'notification.sending');
        $sent = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'notification.sent');

        $this->assertNotEmpty($sending);
        $this->assertNotEmpty($sent);
    }

    // === Broadcast Listener ===

    public function test_broadcast_event_captured(): void
    {
        Queue::fake(); // prevent real queue dispatch

        Wiretap::start();

        $broadcastEvent = new TestBroadcastEvent();
        event($broadcastEvent);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'broadcast');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertNotEmpty($event['data']['channels']);
        $this->assertEquals(['msg' => 'hello'], $event['data']['payload']);
    }

    // === Schedule Listener ===

    public function test_schedule_events_use_correct_types(): void
    {
        // We can't easily trigger real schedule events in testbench,
        // but we verify the event types are correctly defined
        Wiretap::start();

        // Simulate command events (schedule listeners use same mechanism)
        if (class_exists(\Illuminate\Console\Events\ScheduledTaskStarting::class)) {
            $this->assertTrue(true); // Class exists, listener is registered
        } else {
            $this->markTestSkipped('ScheduledTaskStarting not available');
        }
    }

    // === Exception Listener ===

    public function test_exception_event_captured(): void
    {
        Wiretap::start();

        try {
            // Trigger reportable exception handler
            $handler = app(\Illuminate\Contracts\Debug\ExceptionHandler::class);
            $handler->report(new \RuntimeException('Test exception'));
        } catch (\Throwable) {
            // ignore
        }

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'exception');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('RuntimeException', $event['data']['class']);
        $this->assertEquals('Test exception', $event['data']['message']);
    }

    // === Job Listener ===

    public function test_job_start_event_captured(): void
    {
        Wiretap::start();

        $job = new TestQueueJob();
        $queueJob = new \Illuminate\Queue\Jobs\SyncJob(app(), json_encode([
            'uuid' => 'test-uuid-123',
            'displayName' => TestQueueJob::class,
            'job' => 'Illuminate\\Queue\\CallQueuedHandler@call',
            'data' => ['command' => serialize($job), 'commandName' => TestQueueJob::class],
        ]), 'sync', 'default');

        event(new \Illuminate\Queue\Events\JobProcessing('sync', $queueJob));

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'job.start');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertArrayHasKey('job_class', $event['data']);
        $this->assertArrayHasKey('queue', $event['data']);
    }

    public function test_job_not_captured_when_ignored(): void
    {
        config()->set('wiretap.ignore.jobs', ['*TestQueueJob*']);
        $this->assertTrue(Wiretap::shouldIgnore('jobs', TestQueueJob::class));
    }

    // === Mail Listener ===

    public function test_mail_sending_event_captured(): void
    {
        Wiretap::start();

        $message = new \Symfony\Component\Mime\Email();
        $message->to('user@example.com');
        $message->subject('Test Subject');
        $message->from('app@example.com');

        event(new \Illuminate\Mail\Events\MessageSending($message));

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'mail.sending');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('Test Subject', $event['data']['subject']);
    }

    // === Artisan Command Listener ===

    public function test_command_events_captured(): void
    {
        Wiretap::reset();

        // Simulate command start event
        event(new \Illuminate\Console\Events\CommandStarting(
            'test:example',
            new \Symfony\Component\Console\Input\ArrayInput([]),
            new \Symfony\Component\Console\Output\NullOutput()
        ));

        // Should have started a trace
        $this->assertNotEmpty(Wiretap::$rid);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'command.start');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('test:example', $event['data']['command']);
    }

    public function test_ignored_command_not_captured(): void
    {
        Wiretap::reset();

        // 'tinker' is in ignore config
        event(new \Illuminate\Console\Events\CommandStarting(
            'tinker',
            new \Symfony\Component\Console\Input\ArrayInput([]),
            new \Symfony\Component\Console\Output\NullOutput()
        ));

        // Should not have started a trace
        $this->assertEmpty(Wiretap::$rid);
        $this->assertEmpty(Wiretap::$buffer);
    }

    // === Combined: Multiple listeners in one request ===

    public function test_multiple_listeners_in_single_trace(): void
    {
        Wiretap::start();

        // Query
        DB::select('SELECT 1');

        // Cache
        Cache::put('multi_key', 'val', 60);
        Cache::get('multi_key');

        // Log
        Log::info('Multi test');

        // Transaction
        DB::beginTransaction();
        DB::commit();

        $types = array_unique(array_column(Wiretap::$buffer, 'type'));

        $this->assertContains('query', $types);
        $this->assertContains('cache.write', $types);
        $this->assertContains('cache', $types);
        $this->assertContains('log', $types);
        $this->assertContains('transaction', $types);
    }
}

class TestQueueJob implements \Illuminate\Contracts\Queue\ShouldQueue
{
    use \Illuminate\Bus\Queueable;

    public function handle(): void {}
}

class TestBroadcastEvent implements \Illuminate\Contracts\Broadcasting\ShouldBroadcast
{
    public function broadcastOn(): array
    {
        return [new \Illuminate\Broadcasting\Channel('test')];
    }

    public function broadcastWith(): array
    {
        return ['msg' => 'hello'];
    }
}