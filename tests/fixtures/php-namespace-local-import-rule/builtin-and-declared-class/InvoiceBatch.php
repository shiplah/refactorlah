<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function __construct(private string $name) {}
}
