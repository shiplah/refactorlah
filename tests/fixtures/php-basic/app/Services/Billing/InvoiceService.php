<?php

declare(strict_types=1);

namespace App\Services\Billing;

final class InvoiceService
{
    public function invoiceNumber(): string
    {
        return 'INV-001';
    }
}
