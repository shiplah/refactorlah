<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Stmt\UseUse;
use PhpParser\Node\Stmt\Use_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

final class UseStatementReplacementWorker implements ReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<Use_> $useStatements */
        $useStatements = $finder->findInstanceOf($context->ast, Use_::class);

        $replacements = [];
        foreach ($useStatements as $useStatement) {
            if ($useStatement instanceof \PhpParser\Node\Stmt\GroupUse) {
                continue;
            }

            foreach ($useStatement->uses as $useUse) {
                if (!$useUse instanceof UseUse) {
                    continue;
                }
                $resolved = WorkerSupport::resolvedName($useUse->name);
                if (null === $resolved) {
                    $resolved = $useUse->name->toString();
                }
                $mapping = $analysisContext->findByOldSymbol($resolved);
                if (null === $mapping) {
                    continue;
                }

                $replacement = WorkerSupport::createReplacement(
                    $context,
                    $useUse->name,
                    $mapping->newSymbol,
                    'php-use-statement',
                    $this->name(),
                );
                if (null !== $replacement) {
                    $replacements[] = $replacement;
                }
            }
        }

        return $replacements;
    }
}
