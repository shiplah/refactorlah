<?php
namespace App\Services\Items;

final class Helper {}

$createItemService = static function (): object {
    final class ItemService {}

    return new ItemService();
};
