<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig;

use Refactorlah\PhpAdapter\Php\SymbolMapping;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function file_get_contents;
use function is_string;
use function mb_strlen;
use function preg_match_all;
use function preg_quote;
use function str_contains;

final class YamlComponentNamespaceReferenceScanner
{
    /**
     * @param list<string> $files
     * @param list<SymbolMapping> $symbolMappings
     * @return list<Replacement>
     */
    public function scan(string $projectRoot, array $files, array $symbolMappings): array
    {
        if ([] === $symbolMappings) {
            return [];
        }

        $replacements = [];
        foreach ($files as $file) {
            $content = file_get_contents($projectRoot . '/' . $file);
            if (!is_string($content) || !str_contains($content, 'twig_component')) {
                continue;
            }

            foreach ($symbolMappings as $mapping) {
                $replacements = [
                    ...$replacements,
                    ...$this->namespaceDefaultReplacements($file, $content, $mapping),
                ];
            }
        }

        return $replacements;
    }

    /** @return list<Replacement> */
    private function namespaceDefaultReplacements(string $file, string $content, SymbolMapping $mapping): array
    {
        if ($mapping->oldNamespace === $mapping->newNamespace) {
            return [];
        }

        $oldReference = $mapping->oldNamespace . '\\';
        $newReference = $mapping->newNamespace . '\\';
        $pattern = '/([\'"])' . preg_quote($oldReference, '/') . '\1\s*:/';
        if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
            return [];
        }

        $replacements = [];
        foreach ($matches[0] as [$matched, $offset]) {
            $quote = $matched[0];
            $replacements[] = new Replacement(
                file: $file,
                start: $offset,
                end: $offset + mb_strlen($quote . $oldReference . $quote),
                replacement: $quote . $newReference . $quote,
                reason: 'yaml-twig-component-namespace',
                rule: self::class,
            );
        }

        return $replacements;
    }
}
