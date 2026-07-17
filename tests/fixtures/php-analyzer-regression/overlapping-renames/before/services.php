<?php

use App\Module\Lookup\ServiceLookup;
use App\Module\Cache\CacheServiceIndex;

return static function ($services): void {
    $services->set(CacheServiceIndex::class);
    $services->alias(ServiceLookup::class, CacheServiceIndex::class);
};
