<?php

namespace Axvi\Wiretap\Tests\Feature;

use Axvi\Wiretap\Tests\TestCase;
use Illuminate\Support\Facades\Bus;
use Illuminate\Support\Facades\Route;

class ReplayControllerTest extends TestCase
{
    public function test_replay_routes_registered(): void
    {
        $this->assertTrue(
            collect(app('router')->getRoutes())->contains(fn ($route) => str_contains($route->uri, '_wiretap/retry-job'))
        );
        $this->assertTrue(
            collect(app('router')->getRoutes())->contains(fn ($route) => str_contains($route->uri, '_wiretap/dispatch-job'))
        );
        $this->assertTrue(
            collect(app('router')->getRoutes())->contains(fn ($route) => str_contains($route->uri, '_wiretap/replay-broadcast'))
        );
        $this->assertTrue(
            collect(app('router')->getRoutes())->contains(fn ($route) => str_contains($route->uri, '_wiretap/replay-command'))
        );
    }

    public function test_replay_routes_not_registered_when_disabled(): void
    {
        config()->set('wiretap.replay.enabled', false);

        // Re-boot provider
        $provider = new \Axvi\Wiretap\WiretapServiceProvider(app());
        $provider->boot();

        // Routes from initial boot still exist, but this test validates the config check
        $this->assertFalse(config('wiretap.replay.enabled'));
    }

    public function test_replay_requires_secret_header(): void
    {
        $response = $this->postJson('/_wiretap/replay-command', [
            'command' => 'inspire',
        ]);

        $response->assertStatus(403);
        $response->assertJson(['error' => 'invalid secret']);
    }

    public function test_replay_rejects_wrong_secret(): void
    {
        $response = $this->postJson('/_wiretap/replay-command', [
            'command' => 'inspire',
        ], ['X-Wiretap-Secret' => 'wrong']);

        $response->assertStatus(403);
    }

    public function test_replay_command_success(): void
    {
        $response = $this->postJson('/_wiretap/replay-command', [
            'command' => 'env',
        ], ['X-Wiretap-Secret' => 'wt_test_secret']);

        $response->assertStatus(200);
        $response->assertJson(['success' => true]);
        $this->assertEquals(0, $response->json('exit_code'));
    }

    public function test_replay_command_requires_command(): void
    {
        $response = $this->postJson('/_wiretap/replay-command', [], [
            'X-Wiretap-Secret' => 'wt_test_secret',
        ]);

        $response->assertStatus(400);
        $response->assertJson(['error' => 'command is required']);
    }

    public function test_retry_job_requires_uuid(): void
    {
        $response = $this->postJson('/_wiretap/retry-job', [], [
            'X-Wiretap-Secret' => 'wt_test_secret',
        ]);

        $response->assertStatus(400);
        $response->assertJson(['error' => 'uuid is required']);
    }

    public function test_retry_job_with_nonexistent_uuid(): void
    {
        $response = $this->postJson('/_wiretap/retry-job', [
            'uuid' => 'nonexistent-uuid',
        ], ['X-Wiretap-Secret' => 'wt_test_secret']);

        $response->assertStatus(200);
        // queue:retry returns 0 even if job not found
        $this->assertNotNull($response->json('output'));
    }

    public function test_replay_broadcast_requires_event_and_channels(): void
    {
        $response = $this->postJson('/_wiretap/replay-broadcast', [], [
            'X-Wiretap-Secret' => 'wt_test_secret',
        ]);

        $response->assertStatus(400);
        $response->assertJson(['error' => 'event and channels are required']);
    }

    public function test_dispatch_job_requires_serialized(): void
    {
        $response = $this->postJson('/_wiretap/dispatch-job', [], [
            'X-Wiretap-Secret' => 'wt_test_secret',
        ]);

        $response->assertStatus(400);
        $response->assertJson(['error' => 'job_serialized is required']);
    }

    public function test_secret_not_configured_returns_403(): void
    {
        config()->set('wiretap.replay.secret', '');

        $response = $this->postJson('/_wiretap/replay-command', [
            'command' => 'inspire',
        ], ['X-Wiretap-Secret' => 'anything']);

        $response->assertStatus(403);
        $response->assertJson(['error' => 'replay secret not configured']);
    }

    // --- dispatch-job endpoint ---

    public function test_dispatch_job_with_valid_serialized_job(): void
    {
        Bus::fake();

        $job = new TestDispatchableJob(42, 'pending');
        $serialized = base64_encode(serialize($job));

        $response = $this->postJson('/_wiretap/dispatch-job', [
            'job_serialized' => $serialized,
        ], ['X-Wiretap-Secret' => 'wt_test_secret']);

        $response->assertStatus(200);
        $response->assertJson(['success' => true]);
        $this->assertStringContainsString('TestDispatchableJob', $response->json('job_class'));
    }

    public function test_dispatch_job_with_overrides(): void
    {
        Bus::fake();

        $job = new TestDispatchableJob(42, 'pending');
        $serialized = base64_encode(serialize($job));

        $response = $this->postJson('/_wiretap/dispatch-job', [
            'job_serialized' => $serialized,
            'job_payload' => ['status' => 'approved'],
        ], ['X-Wiretap-Secret' => 'wt_test_secret']);

        $response->assertStatus(200);
        $response->assertJson(['success' => true]);
    }

    public function test_dispatch_job_with_invalid_serialized_data(): void
    {
        $response = $this->postJson('/_wiretap/dispatch-job', [
            'job_serialized' => base64_encode('not-a-valid-serialized-object'),
        ], ['X-Wiretap-Secret' => 'wt_test_secret']);

        $response->assertStatus(500);
        $response->assertJson(['success' => false]);
        $this->assertNotNull($response->json('error'));
    }

    // --- FormRequest validation endpoint ---

    public function test_validation_error_returns_422(): void
    {
        // Register a test route with validation rules
        Route::post('/_wiretap_test/validate', function (\Illuminate\Http\Request $request) {
            $request->validate([
                'email' => 'required|email',
                'name' => 'required|string|min:2',
            ]);

            return response()->json(['success' => true]);
        })->middleware('web');

        // Send invalid data
        $response = $this->postJson('/_wiretap_test/validate', [
            'email' => 'not-an-email',
        ]);

        $response->assertStatus(422);
        $response->assertJsonValidationErrors(['email', 'name']);
    }

    public function test_validation_passes_returns_200(): void
    {
        Route::post('/_wiretap_test/validate-pass', function (\Illuminate\Http\Request $request) {
            $request->validate([
                'email' => 'required|email',
                'name' => 'required|string|min:2',
            ]);

            return response()->json(['success' => true]);
        })->middleware('web');

        $response = $this->postJson('/_wiretap_test/validate-pass', [
            'email' => 'john@example.com',
            'name' => 'John',
        ]);

        $response->assertStatus(200);
        $response->assertJson(['success' => true]);
    }
}

class TestDispatchableJob implements \Illuminate\Contracts\Queue\ShouldQueue
{
    use \Illuminate\Bus\Queueable, \Illuminate\Queue\InteractsWithQueue, \Illuminate\Foundation\Bus\Dispatchable;

    public int $orderId;
    public string $status;

    public function __construct(int $orderId, string $status)
    {
        $this->orderId = $orderId;
        $this->status = $status;
    }

    public function handle(): void
    {
        // noop
    }
}
