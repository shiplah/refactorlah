<?php
namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceLineCollection;

final readonly class InvoiceBatch
{
    public function __construct(
        private InvoiceFilter $range,
        private InvoiceLineCollection $documents,
    ) {}
}
