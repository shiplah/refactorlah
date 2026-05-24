<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Replacement;

use Refactorlah\PhpAdapter\Warning\Warning;

final class ReplacementScanResult
{
    /**
     * @param list<Replacement> $replacements
     * @param list<Warning> $warnings
     */
    public function __construct(
        public readonly array $replacements,
        public readonly array $warnings,
    ) {}
}
