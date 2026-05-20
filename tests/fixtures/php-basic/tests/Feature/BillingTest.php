<?php

declare(strict_types=1);

namespace Tests\Feature;

use App\Services\Billing\InvoiceService;

final class BillingTest
{
    /** @param App\Services\Billing\InvoiceService $service */
    public function testInvoice(InvoiceService $service): \App\Services\Billing\InvoiceService
    {
        return new InvoiceService();
    }
}
