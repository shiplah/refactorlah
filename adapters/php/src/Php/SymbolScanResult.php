<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use Refactorlah\PhpAdapter\Warning\Warning;

final class SymbolScanResult
{
    /**
     * @param list<SymbolMapping> $symbolMappings
     * @param list<Warning> $warnings
     */
    public function __construct(
        public readonly array $symbolMappings,
        public readonly array $warnings,
    ) {}
}
