<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node\Scalar\String_;
use PhpParser\Node\Stmt\GroupUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Warning\Warning;

use function array_merge;
use function mb_substr;
use function mb_substr_count;
use function str_contains;

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
            new \Refactorlah\PhpAdapter\Php\Rules\ClassDeclarationReplacementRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\NamespaceLocalDependencyImportRule(),
            new \Refactorlah\PhpAdapter\Php\Rules\UseStatementReplacementRule(),
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

        /** @var list<String_> $strings */
        $strings = $finder->findInstanceOf($context->ast, String_::class);
        foreach ($strings as $string) {
            foreach ($analysisContext->symbolMappings as $mapping) {
                if (!str_contains($string->value, $mapping->oldSymbol)) {
                    continue;
                }

                $offset = $string->getStartFilePos();
                $warnings[] = new Warning(
                    message: 'String literal references a moved PHP symbol; not changed.',
                    file: $context->path,
                    line: $offset < 0 ? $string->getStartLine() : mb_substr_count(mb_substr($context->content, 0, $offset), "\n") + 1,
                );
                break;
            }
        }

        return $warnings;
    }
}
