<?php
namespace App\Consumer;

use App\Source\Item;

final class Consumer
{
    public function service(): Item
    {
        return new Item();
    }
}
