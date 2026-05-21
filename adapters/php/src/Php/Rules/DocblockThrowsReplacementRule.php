<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class DocblockThrowsReplacementRule implements \Refactorlah\PhpAdapter\Php\Rules\ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        return \Refactorlah\PhpAdapter\Php\RuleSupport::docblockTagReplacements(
            $context,
            'throws',
            $analysisContext,
            'php-docblock-throws',
            $this->name(),
        );
    }
}
