<?php

namespace Axvi\Wiretap;

use Illuminate\Console\Command;
use Illuminate\Support\Str;

class GenerateReplaySecretCommand extends Command
{
    protected $signature = 'wiretap:replay-secret
                            {--show : Display the secret instead of modifying files}
                            {--force : Force the operation to run in production}';

    protected $description = 'Generate a Wiretap replay secret key';

    public function handle(): int
    {
        $secret = 'wt_' . Str::random(32);

        if ($this->option('show')) {
            $this->components->info($secret);
            return 0;
        }

        if (! $this->option('force') && app()->environment('production')) {
            $this->components->error('Use --force to generate a secret in production.');
            return 1;
        }

        $this->setEnvValue('WIRETAP_REPLAY_SECRET', $secret);
        $this->setEnvValue('WIRETAP_REPLAY_ENABLED', 'true');

        $this->components->info('Wiretap replay secret set successfully.');
        $this->components->twoColumnDetail('Secret', $secret);
        $this->newLine();
        $this->components->warn('Make sure the same secret is set in your Wiretap Go server config (WIRETAP_REPLAY_SECRET).');

        return 0;
    }

    protected function setEnvValue(string $key, string $value): void
    {
        $path = app()->environmentFilePath();

        if (! file_exists($path)) {
            file_put_contents($path, '');
        }

        $content = file_get_contents($path);

        if (preg_match("/^{$key}=.*/m", $content)) {
            $content = preg_replace("/^{$key}=.*/m", "{$key}={$value}", $content);
        } else {
            $content = rtrim($content, "\n") . "\n{$key}={$value}\n";
        }

        file_put_contents($path, $content);
    }
}