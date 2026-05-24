<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Core;

use Refactorlah\PhpAdapter\Config\PathMapping;
use Refactorlah\PhpAdapter\Config\PathMappingCollection;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function file_get_contents;
use function is_string;
use function mb_strlen;
use function preg_match_all;
use function preg_quote;
use function str_contains;

final class YamlAssetMapperPathReferenceScanner
{
    /**
     * @param list<string> $files
     * @return list<Replacement>
     */
    public function scan(string $projectRoot, array $files, PathMappingCollection $pathMappings): array
    {
        if ($pathMappings->isEmpty()) {
            return [];
        }

        $replacements = [];
        foreach ($files as $file) {
            $content = file_get_contents($projectRoot . '/' . $file);
            if (!is_string($content) || !str_contains($content, 'asset_mapper')) {
                continue;
            }

            foreach ($pathMappings as $mapping) {
                $replacements = [
                    ...$replacements,
                    ...$this->assetMapperPathReplacements($file, $content, $mapping),
                ];
            }
        }

        return $replacements;
    }

    /** @return list<Replacement> */
    private function assetMapperPathReplacements(string $file, string $content, PathMapping $mapping): array
    {
        if ('project-path-directory' !== $mapping->kind) {
            return [];
        }

        $oldReference = $mapping->oldReference;
        $newReference = $mapping->newReference;
        $pattern = '/^(\s*-\s*)([\'"])' . preg_quote($oldReference, '/') . '\2\s*$/m';
        if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
            return [];
        }

        $replacements = [];
        foreach ($matches[2] as [$quote, $offset]) {
            $replacements[] = new Replacement(
                file: $file,
                start: $offset,
                end: $offset + mb_strlen($quote . $oldReference . $quote),
                replacement: $quote . $newReference . $quote,
                reason: 'yaml-asset-mapper-path',
                rule: self::class,
            );
        }

        return $replacements;
    }
}
