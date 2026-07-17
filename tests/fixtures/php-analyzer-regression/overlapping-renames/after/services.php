<?php

use App\Module\Lookup\ServiceLookup;
use App\Module\Cache\CacheServiceLookup;

return static function ($services): void {
    $services->set(CacheServiceLookup::class);
    $services->alias(ServiceLookup::class, CacheServiceLookup::class);
};
