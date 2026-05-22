<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Expr\Instanceof_;
use PhpParser\Node\Expr\New_;
use PhpParser\Node\Expr\StaticCall;
use PhpParser\Node\Expr\StaticPropertyFetch;
use PhpParser\Node\Identifier;
use PhpParser\Node\Name;
use PhpParser\Node\Stmt\Catch_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\RuleSupport;

use function mb_strtolower;

final class ClassNameReferenceReplacementRule implements ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<Name> $names */
        $names = $finder->findInstanceOf($context->ast, Name::class);

        $replacements = [];
        foreach ($names as $name) {
            if (!$this->shouldInspect($name)) {
                continue;
            }

            $resolved = RuleSupport::resolvedName($name);
            if (null === $resolved) {
                continue;
            }

            $mapping = $analysisContext->findByOldSymbol($resolved);
            if (null === $mapping) {
                continue;
            }

            $replacement = RuleSupport::createReplacement(
                $context,
                $name,
                RuleSupport::replacementName($context, $name, $mapping),
                'php-class-name-reference',
                $this->name(),
            );
            if (null !== $replacement) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }

    private function shouldInspect(Name $name): bool
    {
        $original = $name->getAttribute('originalName');
        if ($original instanceof Name && !$original->isUnqualified()) {
            return false;
        }
        if (!$original instanceof Name && !$name->isUnqualified()) {
            return false;
        }

        $parent = $name->getAttribute('parent');

        return match (true) {
            $parent instanceof New_ => $parent->class === $name,
            $parent instanceof StaticCall => $parent->class === $name,
            $parent instanceof StaticPropertyFetch => $parent->class === $name,
            $parent instanceof ClassConstFetch => $parent->class === $name && !$this->isClassNameConstant($parent),
            $parent instanceof Instanceof_ => $parent->class === $name,
            $parent instanceof Catch_ => true,
            default => false,
        };
    }

    private function isClassNameConstant(ClassConstFetch $fetch): bool
    {
        return $fetch->name instanceof Identifier
            && 'class' === mb_strtolower($fetch->name->toString());
    }
}
