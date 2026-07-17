<?php

declare(strict_types=1);

namespace App\Billing\Archive\Domain;

use App\Billing\Domain\InvoiceFilter;
use App\Billing\Domain\InvoiceTotals;

final readonly class InvoiceBatch
{
    public function __construct(
        public string $edition,
        public InvoiceFilter $range,
        public InvoiceTotals $stats,
        public InvoiceLineCollection $documents,
    ) {}
}
