<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

interface ReplacementRule
{
    public function name(): string;

    /** @return list<\Refactorlah\PhpAdapter\Replacement\Replacement> */
    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array;
}
