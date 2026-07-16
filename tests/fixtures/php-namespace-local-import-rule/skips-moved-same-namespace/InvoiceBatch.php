<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function stats(): InvoiceTotals
    {
        return new InvoiceTotals();
    }
}
