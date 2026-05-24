<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use Refactorlah\PhpAdapter\Replacement\Replacement;

use function array_fill;
use function array_merge;
use function array_slice;
use function count;
use function dirname;
use function explode;
use function file_get_contents;
use function implode;
use function is_string;
use function mb_strlen;
use function mb_substr;
use function preg_match_all;
use function preg_quote;
use function str_contains;
use function str_starts_with;
use function trim;

final class StaticImportReferenceScanner
{
    /**
     * @param list<string> $files
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     * @return list<Replacement>
     */
    public function scan(string $projectRoot, array $files, array $moves): array
    {
        $replacements = [];

        foreach ($files as $file) {
            $content = file_get_contents($projectRoot . '/' . $file);
            if (!is_string($content) || '' === $content) {
                continue;
            }

            foreach ($moves as $move) {
                foreach ($this->specifierPairs($file, $move['oldPath'], $move['newPath']) as [$oldSpecifier, $newSpecifier]) {
                    if (!str_contains($content, $oldSpecifier)) {
                        continue;
                    }

                    $replacements = [
                        ...$replacements,
                        ...$this->replacementsForSpecifier($file, $content, $oldSpecifier, $newSpecifier),
                    ];
                }
            }
        }

        return $replacements;
    }

    /**
     * @return list<array{0:string,1:string}>
     */
    private function specifierPairs(string $importingFile, string $oldPath, string $newPath): array
    {
        $oldSpecifier = $this->relativeSpecifier($importingFile, $oldPath);
        $newSpecifier = $this->relativeSpecifier($importingFile, $newPath);
        if ($oldSpecifier === $newSpecifier) {
            return [];
        }

        $pairs = [[$oldSpecifier, $newSpecifier]];
        if (str_starts_with($oldSpecifier, './') && str_starts_with($newSpecifier, './')) {
            $pairs[] = [mb_substr($oldSpecifier, 2), mb_substr($newSpecifier, 2)];
        }

        return $pairs;
    }

    private function relativeSpecifier(string $importingFile, string $targetPath): string
    {
        $fromParts = $this->pathParts(dirname($importingFile));
        $targetParts = $this->pathParts($targetPath);

        $common = 0;
        while (($fromParts[$common] ?? null) !== null
            && ($targetParts[$common] ?? null) !== null
            && $fromParts[$common] === $targetParts[$common]) {
            ++$common;
        }

        $relativeParts = [
            ...array_fill(0, count($fromParts) - $common, '..'),
            ...array_slice($targetParts, $common),
        ];
        $relative = implode('/', $relativeParts);

        if ('' === $relative) {
            return './' . $this->lastPart($targetPath);
        }

        if (!str_starts_with($relative, '..')) {
            return './' . $relative;
        }

        return $relative;
    }

    /** @return list<string> */
    private function pathParts(string $path): array
    {
        $path = trim($path, '/.');
        if ('' === $path) {
            return [];
        }

        return explode('/', $path);
    }

    private function lastPart(string $path): string
    {
        $parts = $this->pathParts($path);

        return $parts[count($parts) - 1] ?? '';
    }

    /**
     * @return list<Replacement>
     */
    private function replacementsForSpecifier(string $file, string $content, string $oldSpecifier, string $newSpecifier): array
    {
        $patterns = [
            '/\bimport\s+(?:[^;\'"]+\s+from\s*)?([\'"])' . preg_quote($oldSpecifier, '/') . '\1/',
            '/\bexport\s+[^;\'"]+\s+from\s*([\'"])' . preg_quote($oldSpecifier, '/') . '\1/',
            '/\bimport\s*\(\s*([\'"])' . preg_quote($oldSpecifier, '/') . '\1\s*\)/',
            '/@import\s+(?:url\(\s*)?([\'"])' . preg_quote($oldSpecifier, '/') . '\1/',
        ];

        $replacements = [];
        foreach ($patterns as $pattern) {
            if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
                continue;
            }

            foreach ($matches[1] as [, $quoteOffset]) {
                $replacements[] = new Replacement(
                    file: $file,
                    start: $quoteOffset + 1,
                    end: $quoteOffset + 1 + mb_strlen($oldSpecifier),
                    replacement: $newSpecifier,
                    reason: 'static-import-path',
                    rule: self::class,
                );
            }
        }

        return $replacements;
    }
}
