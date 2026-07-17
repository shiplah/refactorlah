<?php
namespace App\Module\Source;

use App\Module\Source\ItemList;

final readonly class Container
{
    public function __construct(private ItemList $items) {}
}
