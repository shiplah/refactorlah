<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

interface ReplacementWorker
{
    public function name(): string;

    /** @return list<\Refactorlah\PhpAdapter\Replacement\Replacement> */
    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array;
}
