<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class GroupUseStatementReplacementWorker implements ReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        return [];
    }
}
