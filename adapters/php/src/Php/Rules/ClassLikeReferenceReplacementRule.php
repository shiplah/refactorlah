<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Name;
use PhpParser\Node\Stmt\Class_;
use PhpParser\Node\Stmt\Enum_;
use PhpParser\Node\Stmt\Interface_;
use PhpParser\Node\Stmt\TraitUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\RuleSupport;

use function array_merge;

final class ClassLikeReferenceReplacementRule implements ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        $replacements = [];

        /** @var list<Class_> $classes */
        $classes = $finder->findInstanceOf($context->ast, Class_::class);
        foreach ($classes as $class) {
            $replacements = array_merge($replacements, $this->collectName($context, $analysisContext, $class->extends));
            foreach ($class->implements as $name) {
                $replacements = array_merge($replacements, $this->collectName($context, $analysisContext, $name));
            }
        }

        /** @var list<Interface_> $interfaces */
        $interfaces = $finder->findInstanceOf($context->ast, Interface_::class);
        foreach ($interfaces as $interface) {
            foreach ($interface->extends as $name) {
                $replacements = array_merge($replacements, $this->collectName($context, $analysisContext, $name));
            }
        }

        /** @var list<Enum_> $enums */
        $enums = $finder->findInstanceOf($context->ast, Enum_::class);
        foreach ($enums as $enum) {
            foreach ($enum->implements as $name) {
                $replacements = array_merge($replacements, $this->collectName($context, $analysisContext, $name));
            }
        }

        /** @var list<TraitUse> $traitUses */
        $traitUses = $finder->findInstanceOf($context->ast, TraitUse::class);
        foreach ($traitUses as $traitUse) {
            foreach ($traitUse->traits as $name) {
                $replacements = array_merge($replacements, $this->collectName($context, $analysisContext, $name));
            }
        }

        return $replacements;
    }

    /** @return list<\Refactorlah\PhpAdapter\Replacement\Replacement> */
    private function collectName(PhpFileContext $context, AnalysisContext $analysisContext, ?Name $name): array
    {
        if (null === $name) {
            return [];
        }

        $resolved = RuleSupport::resolvedName($name);
        if (null === $resolved) {
            return [];
        }

        $mapping = $analysisContext->findByOldSymbol($resolved);
        if (null === $mapping) {
            return [];
        }

        $replacement = RuleSupport::createReplacement(
            $context,
            $name,
            RuleSupport::replacementName($context, $name, $mapping),
            'php-class-like-reference',
            $this->name(),
        );

        return null === $replacement ? [] : [$replacement];
    }
}
