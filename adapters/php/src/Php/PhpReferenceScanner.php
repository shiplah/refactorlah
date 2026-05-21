<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node\Stmt\GroupUse;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\Workers\AttributeClassReferenceReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\ClassConstantReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockParamReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockReturnReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockThrowsReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockVarReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\FullyQualifiedClassNameReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\GroupUseStatementReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\MethodParameterTypeReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\MethodReturnTypeReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\NamespaceDeclarationReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\ReplacementWorkerRegistry;
use Refactorlah\PhpAdapter\Php\Workers\TypedPropertyReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\UseStatementReplacementWorker;
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
        $registry = new ReplacementWorkerRegistry(
            new NamespaceDeclarationReplacementWorker(),
            new UseStatementReplacementWorker(),
            new GroupUseStatementReplacementWorker(),
            new FullyQualifiedClassNameReplacementWorker(),
            new ClassConstantReplacementWorker(),
            new DocblockVarReplacementWorker(),
            new DocblockParamReplacementWorker(),
            new DocblockReturnReplacementWorker(),
            new DocblockThrowsReplacementWorker(),
            new AttributeClassReferenceReplacementWorker(),
            new TypedPropertyReplacementWorker(),
            new MethodParameterTypeReplacementWorker(),
            new MethodReturnTypeReplacementWorker(),
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
                $resolved = WorkerSupport::resolvedName($useUse->name);
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
