<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Stmt\Namespace_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class NamespaceDeclarationReplacementRule implements \Refactorlah\PhpAdapter\Php\Rules\ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $mapping = $analysisContext->findByPath($context->path);
        if (null === $mapping || $mapping->oldNamespace === $mapping->newNamespace) {
            return [];
        }

        $finder = new NodeFinder();
        /** @var Namespace_|null $namespace */
        $namespace = $finder->findFirstInstanceOf($context->ast, Namespace_::class);
        if (null === $namespace || null === $namespace->name) {
            return [];
        }

        $replacement = \Refactorlah\PhpAdapter\Php\RuleSupport::createReplacement(
            $context,
            $namespace->name,
            $mapping->newNamespace,
            'php-namespace-declaration',
            $this->name(),
        );

        return null === $replacement ? [] : [$replacement];
    }
}
