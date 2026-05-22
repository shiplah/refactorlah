<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Composer;

use function file_get_contents;
use function is_array;
use function is_file;
use function is_string;
use function json_decode;
use function mb_trim;
use function str_replace;

final class ComposerConfigReader
{
    public function readPsr4Map(string $projectRoot): Psr4Map
    {
        $path = $projectRoot . '/composer.json';
        if (!is_file($path)) {
            throw new \RuntimeException('composer.json is required for PHP adapter analysis');
        }

        $decoded = json_decode((string) file_get_contents($path), true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($decoded)) {
            throw new \RuntimeException('composer.json must decode to an object');
        }

        $autoload = $this->normalizePsr4Block($decoded['autoload'] ?? null);
        $autoloadDev = $this->normalizePsr4Block($decoded['autoload-dev'] ?? null);

        $result = [];
        foreach ([$autoload, $autoloadDev] as $block) {
            foreach ($block as $namespace => $paths) {
                foreach ($paths as $pathValue) {
                    $normalized = $this->normalizePath($pathValue);
                    $result[(string) $namespace][] = $normalized;
                }
            }
        }

        return new Psr4Map($result);
    }

    private function normalizePath(string $path): string
    {
        $normalized = str_replace('\\', '/', $path);
        $normalized = mb_trim($normalized, '/');

        return '' === $normalized ? '.' : $normalized;
    }

    /**
     * @param mixed $block
     * @return array<string,list<string>>
     */
    private function normalizePsr4Block(mixed $block): array
    {
        if (!is_array($block)) {
            return [];
        }

        $psr4 = $block['psr-4'] ?? null;
        if (!is_array($psr4)) {
            return [];
        }

        $normalized = [];
        foreach ($psr4 as $namespace => $paths) {
            if (!is_string($namespace)) {
                continue;
            }

            $pathList = [];
            if (is_array($paths)) {
                foreach ($paths as $path) {
                    if (is_string($path)) {
                        $pathList[] = $path;
                    }
                }
            } elseif (is_string($paths)) {
                $pathList[] = $paths;
            }

            $normalized[$namespace] = $pathList;
        }

        return $normalized;
    }
}
