<?php

declare(strict_types=1);

namespace App\Billing\Domain;

use App\Customer\Domain\CustomerId;

interface InvoiceBatchRepository
{
    public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
}
