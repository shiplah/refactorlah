<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

use function mb_strtolower;

final class ClassConstantReplacementRule implements \Refactorlah\PhpAdapter\Php\Rules\ReplacementRule
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
            if ('class' !== mb_strtolower($fetch->name->toString())) {
                continue;
            }
            if (\Refactorlah\PhpAdapter\Php\RuleSupport::inAttribute($fetch)) {
                continue;
            }
            if (!$fetch->class instanceof \PhpParser\Node\Name) {
                continue;
            }
            $resolved = \Refactorlah\PhpAdapter\Php\RuleSupport::resolvedName($fetch->class);
            if (null === $resolved) {
                continue;
            }
            $mapping = $analysisContext->findByOldSymbol($resolved);
            if (null === $mapping) {
                continue;
            }

            $replacement = \Refactorlah\PhpAdapter\Php\RuleSupport::createReplacement(
                $context,
                $fetch->class,
                \Refactorlah\PhpAdapter\Php\RuleSupport::replacementName($context, $fetch->class, $mapping),
                'php-class-constant',
                $this->name(),
            );
            if (null !== $replacement) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }
}
