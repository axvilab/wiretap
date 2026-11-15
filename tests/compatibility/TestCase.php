<?php

namespace Axvi\Wiretap\Tests\Compatibility;

use Axvi\Wiretap\WiretapServiceProvider;
use Orchestra\Testbench\TestCase as BaseTestCase;

abstract class TestCase extends BaseTestCase
{
    protected function getPackageProviders($app): array
    {
        return [WiretapServiceProvider::class];
    }

    protected function defineEnvironment($app): void
    {
        $app['config']->set('wiretap.enabled', true);
        $app['config']->set('wiretap.replay.enabled', true);
        $app['config']->set('wiretap.replay.secret', 'wt_test_secret');
        $app['config']->set('app.key', 'base64:' . base64_encode(random_bytes(32)));
    }
}