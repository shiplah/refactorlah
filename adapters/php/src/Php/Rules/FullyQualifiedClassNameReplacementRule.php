<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Name\FullyQualified;
use PhpParser\Node\Stmt\UseUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class FullyQualifiedClassNameReplacementRule implements \Refactorlah\PhpAdapter\Php\Rules\ReplacementRule
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
            if (\Refactorlah\PhpAdapter\Php\RuleSupport::inAttribute($name)) {
                continue;
            }
            if (\Refactorlah\PhpAdapter\Php\RuleSupport::isTypeReference($name)) {
                continue;
            }

            $mapping = $analysisContext->findByOldSymbol($name->toString());
            if (null === $mapping) {
                continue;
            }

            $replacement = \Refactorlah\PhpAdapter\Php\RuleSupport::createReplacement(
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
