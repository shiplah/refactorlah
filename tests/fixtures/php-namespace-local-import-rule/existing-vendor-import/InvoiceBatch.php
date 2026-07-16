<?php
namespace App\Billing\Domain;

use Vendor\InvoiceFilter;

final readonly class InvoiceBatch
{
    public function __construct(private InvoiceFilter $range) {}
}
