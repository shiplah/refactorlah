<?php

declare(strict_types=1);

use ShipMonk\ComposerDependencyAnalyser\Config\Configuration;

$config = new Configuration();

return $config
    ->addPathToScan(__DIR__ . '/src', isDev: false)
    ->addPathToScan(__DIR__ . '/bin', isDev: false)
    ->addPathToScan(__DIR__ . '/tests', isDev: true)
    ->addPathToExclude(__DIR__ . '/tests/fixtures')
    ->addPathToScan(__DIR__ . '/.php-cs-fixer.dist.php', isDev: true)
    ->addPathToScan(__DIR__ . '/rector.php', isDev: true)
;
