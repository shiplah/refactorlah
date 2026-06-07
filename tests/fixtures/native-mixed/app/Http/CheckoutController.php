<?php

declare(strict_types=1);

namespace App\Http;

use App\Services\Billing\InvoiceService;

final class CheckoutController
{
    /** @param iterable<InvoiceService> $services @return \App\Services\Billing\InvoiceService */
    public function checkout(InvoiceService $service): \App\Services\Billing\InvoiceService
    {
        $template = $this->render('billing/invoice.html.twig');
        $class = \App\Services\Billing\InvoiceService::class;

        return new InvoiceService();
    }

    private function render(string $template): string
    {
        return $template;
    }
}
