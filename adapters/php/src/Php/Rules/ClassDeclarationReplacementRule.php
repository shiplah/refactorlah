<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node;
use PhpParser\Node\Stmt\Class_;
use PhpParser\Node\Stmt\Enum_;
use PhpParser\Node\Stmt\Interface_;
use PhpParser\Node\Stmt\Trait_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\RuleSupport;

use function mb_strrpos;
use function mb_substr;

final class ClassDeclarationReplacementRule implements ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $mapping = $analysisContext->findByPath($context->path);
        if (null === $mapping) {
            return [];
        }

        $newShortName = $this->shortName($mapping->newSymbol);
        if ($newShortName === $mapping->shortName) {
            return [];
        }

        $finder = new NodeFinder();
        /** @var list<Class_|Interface_|Trait_|Enum_> $symbols */
        $symbols = $finder->find($context->ast, static fn(Node $node): bool => $node instanceof Class_
            || $node instanceof Interface_
            || $node instanceof Trait_
            || $node instanceof Enum_);

        foreach ($symbols as $symbol) {
            if ($symbol->name?->toString() !== $mapping->shortName) {
                continue;
            }

            $replacement = RuleSupport::createReplacement(
                $context,
                $symbol->name,
                $newShortName,
                'php-class-declaration',
                $this->name(),
            );

            return null === $replacement ? [] : [$replacement];
        }

        return [];
    }

    private function shortName(string $symbol): string
    {
        $index = mb_strrpos($symbol, '\\');
        if (false === $index) {
            return $symbol;
        }

        return mb_substr($symbol, $index + 1);
    }
}
