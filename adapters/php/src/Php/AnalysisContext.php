<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

final class AnalysisContext
{
    /**
     * @param array<string,SymbolMapping> $symbolMappings
     */
    public function __construct(
        public readonly array $symbolMappings,
    ) {
    }

    public function findByOldSymbol(string $symbol): ?SymbolMapping
    {
        return $this->symbolMappings[$symbol] ?? null;
    }

    public function findByPath(string $path): ?SymbolMapping
    {
        foreach ($this->symbolMappings as $mapping) {
            if ($mapping->oldPath === $path) {
                return $mapping;
            }
        }

        return null;
    }
}
