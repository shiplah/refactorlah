<?php
namespace App\Consumer;

use App\Source\Item;

final class Consumer
{
    public const SERVICE = \App\Source\Item::class;

    public function service(): \App\Source\Item {}
}
