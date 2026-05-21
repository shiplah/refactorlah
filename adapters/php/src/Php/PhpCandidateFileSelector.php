<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use function array_fill_keys;
use function array_keys;
use function file_get_contents;
use function is_string;
use function sort;
use function str_contains;

final class PhpCandidateFileSelector
{
    /**
     * @param list<string> $files
     * @param list<SymbolMapping> $symbolMappings
     * @param list<string> $movedPhpFiles
     * @return list<string>
     */
    public function select(string $projectRoot, array $files, array $symbolMappings, array $movedPhpFiles): array
    {
        if ([] === $symbolMappings) {
            return [];
        }

        $movedIndex = array_fill_keys($movedPhpFiles, true);
        $needles = $this->needles($symbolMappings);
        $selected = [];

        foreach ($files as $file) {
            if (isset($movedIndex[$file])) {
                $selected[] = $file;
                continue;
            }

            $content = @file_get_contents($projectRoot . '/' . $file);
            if (!is_string($content) || '' === $content) {
                continue;
            }

            foreach ($needles as $needle) {
                if ('' !== $needle && str_contains($content, $needle)) {
                    $selected[] = $file;
                    break;
                }
            }
        }

        sort($selected);

        return $selected;
    }

    /**
     * @param list<SymbolMapping> $symbolMappings
     * @return list<string>
     */
    private function needles(array $symbolMappings): array
    {
        $needles = [];

        foreach ($symbolMappings as $mapping) {
            $needles[$mapping->oldSymbol] = true;
            $needles[$mapping->oldNamespace] = true;
            $needles[$mapping->shortName] = true;
        }

        return array_keys($needles);
    }
}
