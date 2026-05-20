<?php

declare(strict_types=1);

namespace App\Http\Controllers;

use App\Services\Billing\InvoiceService;

#[Template('admin/user/card.html.twig', service: InvoiceService::class)]
final class InvoiceController
{
    /** @var App\Services\Billing\InvoiceService */
    private \App\Services\Billing\InvoiceService $service;

    public function __construct(InvoiceService $service)
    {
        $this->service = $service;
    }

    /** @return App\Services\Billing\InvoiceService */
    public function service(): \App\Services\Billing\InvoiceService
    {
        return $this->service;
    }

    /** @throws App\Services\Billing\InvoiceService */
    public function show(): array
    {
        return [
            'service' => InvoiceService::class,
            'fqcn' => \App\Services\Billing\InvoiceService::class,
            'template' => $this->render('admin/user/card.html.twig'),
            'dynamicTemplate' => $this->render($this->chooseTemplate()),
        ];
    }

    private function render(string $template): string
    {
        return $template;
    }

    private function chooseTemplate(): string
    {
        return 'admin/user/card.html.twig';
    }
}
