<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function __construct(private InvoiceFilter $range) {}

    public function stats(): InvoiceTotals
    {
        return new InvoiceTotals();
    }
}
