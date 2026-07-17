<?php
namespace App\Consumer;

use App\Module\Source\Resolver;

final class Consumer
{
    public function __construct(private Resolver $resolver) {}
}
