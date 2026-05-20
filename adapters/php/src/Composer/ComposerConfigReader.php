<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Composer;

use RuntimeException;

final class ComposerConfigReader
{
    public function readPsr4Map(string $projectRoot): Psr4Map
    {
        $path = $projectRoot . '/composer.json';
        if (!is_file($path)) {
            throw new RuntimeException('composer.json is required for PHP adapter analysis');
        }

        $decoded = json_decode((string) file_get_contents($path), true, flags: JSON_THROW_ON_ERROR);
        $autoload = $decoded['autoload']['psr-4'] ?? [];
        $autoloadDev = $decoded['autoload-dev']['psr-4'] ?? [];

        $result = [];
        foreach ([$autoload, $autoloadDev] as $block) {
            foreach ($block as $namespace => $paths) {
                foreach ((array) $paths as $pathValue) {
                    $normalized = $this->normalizePath((string) $pathValue);
                    $result[(string) $namespace][] = $normalized;
                }
            }
        }

        return new Psr4Map($result);
    }

    private function normalizePath(string $path): string
    {
        $normalized = str_replace('\\', '/', $path);
        $normalized = trim($normalized, '/');

        return $normalized === '' ? '.' : $normalized;
    }
}
