<?php
namespace App\PhpSrcTestStyle\Fixing;

use FilesystemIterator;
use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use RuntimeException;

final readonly class Runner
{
    private const LABELS = ['case'];

    public function paths(string $root): array
    {
        $path = $root . DIRECTORY_SEPARATOR . 'cases';
        if (! glob($path, GLOB_ONLYDIR)) {
            throw new RuntimeException(dirname(__DIR__, 2));
        }

        new RecursiveDirectoryIterator($path, FilesystemIterator::SKIP_DOTS);
        RecursiveIteratorIterator::LEAVES_ONLY;

        return self::LABELS;
    }
}
