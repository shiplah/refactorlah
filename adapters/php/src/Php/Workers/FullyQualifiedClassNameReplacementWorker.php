<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Name\FullyQualified;
use PhpParser\Node\Stmt\UseUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

final class FullyQualifiedClassNameReplacementWorker implements ReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<FullyQualified> $names */
        $names = $finder->findInstanceOf($context->ast, FullyQualified::class);

        $replacements = [];
        foreach ($names as $name) {
            $parent = $name->getAttribute('parent');
            if ($parent instanceof UseUse || $parent instanceof ClassConstFetch) {
                continue;
            }
            if (WorkerSupport::inAttribute($name)) {
                continue;
            }
            if (WorkerSupport::isTypeReference($name)) {
                continue;
            }

            $mapping = $analysisContext->findByOldSymbol($name->toString());
            if (null === $mapping) {
                continue;
            }

            $replacement = WorkerSupport::createReplacement(
                $context,
                $name,
                '\\' . $mapping->newSymbol,
                'php-fully-qualified-class-name',
                $this->name(),
            );
            if (null !== $replacement) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }
}
