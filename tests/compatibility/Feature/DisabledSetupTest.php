<?php

namespace Axvi\Wiretap\Tests\Compatibility\Feature;

use Axvi\Wiretap\Tests\Compatibility\TestCase;
use Axvi\Wiretap\WiretapServiceProvider;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\ServiceProvider;

/**
 * A fresh install has WIRETAP_ENABLED unset, and the README asks to publish the config first.
 */
class DisabledSetupTest extends TestCase
{
    protected function defineEnvironment($app): void
    {
        parent::defineEnvironment($app);
        $app['config']->set('wiretap.enabled', false);
    }

    public function test_config_can_be_published_while_disabled(): void
    {
        $paths = ServiceProvider::pathsToPublish(WiretapServiceProvider::class, 'wiretap-config');

        $this->assertCount(1, $paths);
        $this->assertSame(config_path('wiretap.php'), array_values($paths)[0]);
    }

    public function test_replay_secret_command_is_available_while_disabled(): void
    {
        $this->assertArrayHasKey('wiretap:replay-secret', Artisan::all());
    }
}
