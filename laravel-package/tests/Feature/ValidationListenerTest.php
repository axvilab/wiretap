<?php

namespace Axvi\Wiretap\Tests\Feature;

use Axvi\Wiretap\Tests\TestCase;
use Axvi\Wiretap\Wiretap;

class ValidationListenerTest extends TestCase
{
    protected function defineRoutes($router): void
    {
        $router->post('/_test/validation-basic', function (\Illuminate\Http\Request $request) {
            $request->validate([
                'email' => 'required|email',
                'name' => 'required|string|min:2',
            ]);
            return response()->json(['success' => true]);
        });

        $router->post('/_test/validation-nested', function (\Illuminate\Http\Request $request) {
            $request->validate([
                'user.email' => 'required|email',
                'user.age' => 'required|integer|min:18',
            ]);
            return response()->json(['success' => true]);
        });

        $router->post('/_test/validation-passes', function (\Illuminate\Http\Request $request) {
            $request->validate([
                'name' => 'required|string',
            ]);
            return response()->json(['success' => true]);
        });
    }

    public function test_validation_error_returns_422_with_errors(): void
    {
        $response = $this->postJson('/_test/validation-basic', [
            'email' => 'not-an-email',
        ]);

        $response->assertStatus(422);
        $response->assertJsonValidationErrors(['email', 'name']);
    }

    public function test_nested_validation_error_returns_422(): void
    {
        $response = $this->postJson('/_test/validation-nested', [
            'user' => ['email' => 'bad', 'age' => 5],
        ]);

        $response->assertStatus(422);
        $response->assertJsonValidationErrors(['user.email', 'user.age']);
    }

    public function test_validation_passes_returns_200(): void
    {
        $response = $this->postJson('/_test/validation-passes', [
            'name' => 'John',
        ]);

        $response->assertStatus(200);
        $response->assertJson(['success' => true]);
    }

    public function test_validation_event_not_captured_without_rid(): void
    {
        Wiretap::reset();

        $validator = \Illuminate\Support\Facades\Validator::make(
            ['email' => 'bad'],
            ['email' => 'required|email']
        );

        try {
            $validator->validate();
        } catch (\Illuminate\Validation\ValidationException $e) {
            // renderable handler checks rid — should not buffer
        }

        $validationEvents = array_filter(Wiretap::$buffer, fn ($e) => $e['type'] === 'validation');
        $this->assertEmpty($validationEvents);
    }
}
