<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

final class DocblockVarReplacementWorker implements ReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        return WorkerSupport::docblockTagReplacements(
            $context,
            'var',
            $analysisContext,
            'php-docblock-var',
            $this->name(),
        );
    }
}
