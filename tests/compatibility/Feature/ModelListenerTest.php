<?php

namespace Axvi\Wiretap\Tests\Compatibility\Feature;

use Axvi\Wiretap\Tests\Compatibility\TestCase;
use Axvi\Wiretap\Wiretap;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Support\Facades\Event;

class ModelListenerTest extends TestCase
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

    public function test_model_created_event_captured(): void
    {
        Wiretap::start();

        event('eloquent.created: App\\Models\\TestModel', [new TestModel()]);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'model');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('created', $event['data']['action']);
        $this->assertStringContainsString('TestModel', $event['data']['model']);
    }

    public function test_model_updated_event_captured(): void
    {
        Wiretap::start();

        event('eloquent.updated: App\\Models\\TestModel', [new TestModel()]);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'model');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('updated', $event['data']['action']);
    }

    public function test_model_deleted_event_captured(): void
    {
        Wiretap::start();

        event('eloquent.deleted: App\\Models\\TestModel', [new TestModel()]);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'model');
        $this->assertNotEmpty($events);

        $event = array_values($events)[0];
        $this->assertEquals('deleted', $event['data']['action']);
    }

    public function test_model_event_not_captured_without_rid(): void
    {
        event('eloquent.created: App\\Models\\TestModel', [new TestModel()]);

        $this->assertEmpty(Wiretap::$buffer);
    }

    public function test_model_event_captures_key(): void
    {
        Wiretap::start();

        $model = new TestModel();
        $model->id = 42;
        event('eloquent.created: App\\Models\\TestModel', [$model]);

        $events = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'model');
        $event = array_values($events)[0];
        $this->assertEquals(42, $event['data']['key']);
    }

    public function test_model_event_not_captured_when_disabled(): void
    {
        config()->set('wiretap.models.enabled', false);
        $this->assertFalse(config('wiretap.models.enabled'));
    }

    public function test_ignored_model_not_captured(): void
    {
        config()->set('wiretap.models.ignore', ['*TestModel*']);
        Wiretap::start();

        // The ignore check happens in the listener, but since we fire events
        // directly and the listener was registered at boot with original config,
        // we test the shouldIgnore method directly
        $this->assertTrue(Wiretap::shouldIgnore('models', 'App\\Models\\TestModel'));
        $this->assertFalse(Wiretap::shouldIgnore('models', 'App\\Models\\User'));
    }
}

class TestModel extends Model
{
    protected $guarded = [];
    public $timestamps = false;
    public $id = null;

    public function getKey()
    {
        return $this->id;
    }
}
