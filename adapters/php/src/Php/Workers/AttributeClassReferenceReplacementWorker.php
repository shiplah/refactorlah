<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

use function mb_strtolower;

final class AttributeClassReferenceReplacementWorker implements ReplacementWorker
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
            if (!WorkerSupport::inAttribute($fetch)) {
                continue;
            }
            if ('class' !== mb_strtolower($fetch->name->toString())) {
                continue;
            }
            if (!$fetch->class instanceof \PhpParser\Node\Name) {
                continue;
            }
            $resolved = WorkerSupport::resolvedName($fetch->class);
            if (null === $resolved) {
                continue;
            }
            $mapping = $analysisContext->findByOldSymbol($resolved);
            if (null === $mapping) {
                continue;
            }

            $replacement = WorkerSupport::createReplacement(
                $context,
                $fetch->class,
                WorkerSupport::replacementName($context, $fetch->class, $mapping),
                'php-attribute-class-reference',
                $this->name(),
            );
            if (null !== $replacement) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }
}
