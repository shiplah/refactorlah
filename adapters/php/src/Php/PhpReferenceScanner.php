<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node\Stmt\GroupUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Warning\Warning;

use function array_merge;

final class PhpReferenceScanner
{
    /**
     * @param list<PhpFileContext> $contexts
     * @return array{0:list<Replacement>,1:list<Warning>}
     */
    public function scan(array $contexts, AnalysisContext $analysisContext): array
    {
        $registry = new \Refactorlah\PhpAdapter\Php\Rules\ReplacementRuleRegistry(
            new \Refactorlah\PhpAdapter\Php\Rules\NamespaceDeclarationReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\UseStatementReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\GroupUseStatementReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\FullyQualifiedClassNameReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\ClassConstantReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\DocblockVarReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\DocblockParamReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\DocblockReturnReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\DocblockThrowsReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\AttributeClassReferenceReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\TypedPropertyReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\MethodParameterTypeReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\MethodReturnTypeReplacementRule(),
        );

        $replacements = [];
        $warnings = [];
        foreach ($contexts as $context) {
            $replacements = array_merge($replacements, $registry->collect($context, $analysisContext));
            $warnings = array_merge($warnings, $this->collectWarnings($context, $analysisContext));
        }

        return [$replacements, $warnings];
    }

    /** @return list<Warning> */
    private function collectWarnings(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        $warnings = [];

        /** @var list<GroupUse> $groupUses */
        $groupUses = $finder->findInstanceOf($context->ast, GroupUse::class);
        foreach ($groupUses as $groupUse) {
            foreach ($groupUse->uses as $useUse) {
                $resolved = \Refactorlah\PhpAdapter\Php\RuleSupport::resolvedName($useUse->name);
                if (null === $resolved || null === $analysisContext->findByOldSymbol($resolved)) {
                    continue;
                }

                $warnings[] = new Warning(
                    message: 'Group use statement references a moved symbol; skipped conservatively.',
                    file: $context->path,
                    line: $groupUse->getStartLine(),
                );
            }
        }

        return $warnings;
    }
}
