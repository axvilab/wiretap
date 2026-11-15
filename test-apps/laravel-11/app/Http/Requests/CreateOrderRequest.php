<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class CreateOrderRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    public function rules(): array
    {
        return [
            'product' => 'required|string|min:3',
            'quantity' => 'required|integer|min:1',
            'email' => 'required|email',
        ];
    }
}
