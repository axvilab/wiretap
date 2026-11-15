<?php

namespace App\Mail;

use Illuminate\Bus\Queueable;
use Illuminate\Mail\Mailable;
use Illuminate\Mail\Mailables\Content;
use Illuminate\Mail\Mailables\Envelope;
use Illuminate\Queue\SerializesModels;

class OrderConfirmation extends Mailable
{
    use Queueable, SerializesModels;

    public function __construct(
        public int $orderId,
        public float $total,
        public array $items,
    ) {}

    public function envelope(): Envelope
    {
        return new Envelope(
            subject: "Order #{$this->orderId} Confirmed",
        );
    }

    public function content(): Content
    {
        return new Content(
            htmlString: self::renderHtml($this->orderId, $this->total, $this->items),
        );
    }

    public function attachments(): array
    {
        // Generate a test CSV invoice
        $csv = "Item,Qty,Price\n";
        foreach ($this->items as $item) {
            $csv .= "{$item['name']},{$item['qty']},{$item['price']}\n";
        }
        $csv .= "Total,,{$this->total}\n";

        // Generate a test TXT receipt
        $txt = "RECEIPT — Order #{$this->orderId}\n";
        $txt .= str_repeat('=', 30) . "\n";
        foreach ($this->items as $item) {
            $txt .= sprintf("%-20s x%d  $%s\n", $item['name'], $item['qty'], $item['price']);
        }
        $txt .= str_repeat('-', 30) . "\n";
        $txt .= sprintf("TOTAL: $%.2f\n", $this->total);

        return [
            \Illuminate\Mail\Mailables\Attachment::fromData(fn () => $csv, "invoice-{$this->orderId}.csv")
                ->withMime('text/csv'),
            \Illuminate\Mail\Mailables\Attachment::fromData(fn () => $txt, "receipt-{$this->orderId}.txt")
                ->withMime('text/plain'),
        ];
    }

    public static function renderHtml(int $orderId, float $total, array $items): string
    {
        $itemsHtml = '';
        foreach ($items as $item) {
            $itemsHtml .= <<<HTML
            <tr>
                <td style="padding: 12px 16px; border-bottom: 1px solid #e5e7eb;">{$item['name']}</td>
                <td style="padding: 12px 16px; border-bottom: 1px solid #e5e7eb; text-align: center;">{$item['qty']}</td>
                <td style="padding: 12px 16px; border-bottom: 1px solid #e5e7eb; text-align: right;">\${$item['price']}</td>
            </tr>
            HTML;
        }

        return <<<HTML
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
        </head>
        <body style="margin: 0; padding: 0; background-color: #f3f4f6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;">
            <div style="max-width: 600px; margin: 0 auto; padding: 40px 20px;">
                <!-- Header -->
                <div style="background: linear-gradient(135deg, #059669, #10b981); border-radius: 12px 12px 0 0; padding: 32px; text-align: center;">
                    <h1 style="color: white; margin: 0; font-size: 24px;">Order Confirmed!</h1>
                    <p style="color: rgba(255,255,255,0.85); margin: 8px 0 0; font-size: 14px;">Thank you for your purchase</p>
                </div>

                <!-- Content -->
                <div style="background: white; padding: 32px; border-radius: 0 0 12px 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.05);">
                    <p style="color: #374151; font-size: 16px; margin: 0 0 8px;">Hi there,</p>
                    <p style="color: #6b7280; font-size: 14px; margin: 0 0 24px;">
                        Your order <strong style="color: #059669;">#{$orderId}</strong> has been confirmed and is being processed.
                    </p>

                    <!-- Order Items -->
                    <table style="width: 100%; border-collapse: collapse; margin-bottom: 24px;">
                        <thead>
                            <tr style="background-color: #f9fafb;">
                                <th style="padding: 12px 16px; text-align: left; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Item</th>
                                <th style="padding: 12px 16px; text-align: center; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Qty</th>
                                <th style="padding: 12px 16px; text-align: right; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Price</th>
                            </tr>
                        </thead>
                        <tbody style="font-size: 14px; color: #374151;">
                            $itemsHtml
                        </tbody>
                    </table>

                    <!-- Total -->
                    <div style="background-color: #f9fafb; border-radius: 8px; padding: 16px; display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-size: 14px; color: #6b7280;">Total</span>
                        <span style="font-size: 20px; font-weight: bold; color: #059669;">\${$total}</span>
                    </div>

                    <!-- CTA -->
                    <div style="text-align: center; margin-top: 32px;">
                        <a href="#" style="display: inline-block; background-color: #059669; color: white; padding: 12px 32px; border-radius: 8px; text-decoration: none; font-size: 14px; font-weight: 600;">
                            Track Your Order
                        </a>
                    </div>

                    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 32px 0 16px;">

                    <p style="color: #9ca3af; font-size: 12px; text-align: center; margin: 0;">
                        This is a test email from Wiretap test application.
                    </p>
                </div>
            </div>
        </body>
        </html>
        HTML;
    }
}
