<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node;
use PhpParser\Node\Name;
use PhpParser\Node\NullableType;
use PhpParser\Node\UnionType;
use PhpParser\Node\IntersectionType;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\WorkerSupport;

abstract class AbstractTypeReplacementWorker implements ReplacementWorker
{
    /**
     * @return list<\Refactorlah\PhpAdapter\Replacement\Replacement>
     */
    protected function collectTypeReplacements(
        PhpFileContext $context,
        AnalysisContext $analysisContext,
        Node|string|null $type,
        string $reason,
    ): array {
        if ($type === null || is_string($type)) {
            return [];
        }

        if ($type instanceof NullableType) {
            return $this->collectTypeReplacements($context, $analysisContext, $type->type, $reason);
        }

        if ($type instanceof UnionType || $type instanceof IntersectionType) {
            $replacements = [];
            foreach ($type->types as $nestedType) {
                $replacements = array_merge($replacements, $this->collectTypeReplacements($context, $analysisContext, $nestedType, $reason));
            }
            return $replacements;
        }

        if (!$type instanceof Name) {
            return [];
        }

        $resolved = WorkerSupport::resolvedName($type);
        if ($resolved === null) {
            return [];
        }

        $mapping = $analysisContext->findByOldSymbol($resolved);
        if ($mapping === null) {
            return [];
        }

        $replacement = WorkerSupport::createReplacement(
            $context,
            $type,
            '\\' . $mapping->newSymbol,
            $reason,
            $this->name(),
        );

        return $replacement === null ? [] : [$replacement];
    }
}
