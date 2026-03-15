<?php

namespace Axvi\Wiretap\Tests\Feature;

use Axvi\Wiretap\Tests\TestCase;

class GenerateReplaySecretCommandTest extends TestCase
{
    public function test_command_generates_secret(): void
    {
        $this->artisan('wiretap:replay-secret')
            ->assertExitCode(0)
            ->expectsOutputToContain('Wiretap replay secret set successfully');
    }

    public function test_command_show_flag(): void
    {
        $this->artisan('wiretap:replay-secret', ['--show' => true])
            ->assertExitCode(0);
    }

    public function test_generated_secret_has_prefix(): void
    {
        // Run with --show to get the secret without writing to .env
        $this->artisan('wiretap:replay-secret', ['--show' => true])
            ->expectsOutputToContain('wt_');
    }
}
