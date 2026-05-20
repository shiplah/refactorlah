<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

final class ClassConstantReplacementWorker implements ReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<ClassConstFetch> $fetches */
        $fetches = $finder->findInstanceOf($context->ast, ClassConstFetch::class);

        $replacements = [];
        foreach ($fetches as $fetch) {
            if (strtolower($fetch->name->toString()) !== 'class') {
                continue;
            }
            if (WorkerSupport::inAttribute($fetch)) {
                continue;
            }
            if (!$fetch->class instanceof \PhpParser\Node\Name) {
                continue;
            }
            $resolved = WorkerSupport::resolvedName($fetch->class);
            if ($resolved === null) {
                continue;
            }
            $mapping = $analysisContext->findByOldSymbol($resolved);
            if ($mapping === null) {
                continue;
            }

            $replacement = WorkerSupport::createReplacement(
                $context,
                $fetch->class,
                '\\' . $mapping->newSymbol,
                'php-class-constant',
                $this->name(),
            );
            if ($replacement !== null) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }
}
