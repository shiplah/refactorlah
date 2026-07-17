<?php
namespace App\Services\Items;

final class ItemService {}

function createItemService(): object
{
    class NestedItemService {}

    return new NestedItemService();
}
