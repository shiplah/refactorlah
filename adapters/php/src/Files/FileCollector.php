<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Files;

use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use SplFileInfo;

final class FileCollector
{
    /**
     * @param list<string> $extensions
     * @return list<string>
     */
    public function collect(string $projectRoot, array $extensions): array
    {
        $collected = [];
        $allowed = array_fill_keys(array_map('strtolower', $extensions), true);
        $iterator = new RecursiveIteratorIterator(
            new RecursiveDirectoryIterator($projectRoot, RecursiveDirectoryIterator::SKIP_DOTS)
        );

        /** @var SplFileInfo $file */
        foreach ($iterator as $file) {
            if (!$file->isFile()) {
                continue;
            }

            $relative = str_replace('\\', '/', substr($file->getPathname(), strlen($projectRoot) + 1));
            if ($this->isIgnored($relative)) {
                continue;
            }

            $extension = strtolower($file->getExtension());
            if (!isset($allowed[$extension])) {
                continue;
            }

            $collected[] = $relative;
        }

        sort($collected);
        return $collected;
    }

    private function isIgnored(string $path): bool
    {
        foreach ([
            '.git/',
            'vendor/',
            'node_modules/',
            'var/',
            'storage/framework/',
            'bootstrap/cache/',
            'build/',
            'dist/',
            'coverage/',
        ] as $prefix) {
            if (str_starts_with($path, $prefix)) {
                return true;
            }
        }

        return false;
    }
}
