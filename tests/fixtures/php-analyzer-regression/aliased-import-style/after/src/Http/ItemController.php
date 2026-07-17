<?php

declare(strict_types=1);

namespace App\Http;

use App\Domain\Items\DomainItemService as ItemApi;

final class ItemController
{
    public function service(): ItemApi
    {
        return new ItemApi();
    }
}
