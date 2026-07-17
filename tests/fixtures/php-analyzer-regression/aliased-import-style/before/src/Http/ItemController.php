<?php

declare(strict_types=1);

namespace App\Http;

use App\Services\Items\ItemService as ItemApi;

final class ItemController
{
    public function service(): ItemApi
    {
        return new ItemApi();
    }
}
