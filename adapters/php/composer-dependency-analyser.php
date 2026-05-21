<?php

declare(strict_types=1);

use ShipMonk\ComposerDependencyAnalyser\Config\Configuration;
use ShipMonk\ComposerDependencyAnalyser\Config\ErrorType;

$config = new Configuration();

return $config
    ->addPathToScan(__DIR__ . '/src', isDev: false)
    ->addPathToScan(__DIR__ . '/bin', isDev: false)
    ->addPathToScan(__DIR__ . '/tests', isDev: true)
    ->addPathToScan(__DIR__ . '/.php-cs-fixer.dist.php', isDev: true)
    ->addPathToScan(__DIR__ . '/rector.php', isDev: true)
    ->ignoreErrorsOnPackage('phpstan/phpstan', [ErrorType::UNUSED_DEPENDENCY])
    ->ignoreErrorsOnPackage('shipmonk/dead-code-detector', [ErrorType::UNUSED_DEPENDENCY])
;
