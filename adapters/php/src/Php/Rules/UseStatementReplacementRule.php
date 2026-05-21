<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Stmt\GroupUse;
use PhpParser\Node\Stmt\UseUse;
use PhpParser\Node\Stmt\Use_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function implode;
use function is_int;
use function mb_strlen;

final class UseStatementReplacementRule implements \Refactorlah\PhpAdapter\Php\Rules\ReplacementRule
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

        $effectiveNamespace = \Refactorlah\PhpAdapter\Php\RuleSupport::effectiveNamespace($context, $analysisContext);
        $replacements = [];
        foreach ($useStatements as $useStatement) {
            if ($useStatement instanceof GroupUse) {
                continue;
            }

            $updatedUses = [];
            $changed = false;

            foreach ($useStatement->uses as $useUse) {
                if (!$useUse instanceof UseUse) {
                    continue;
                }
                $resolved = \Refactorlah\PhpAdapter\Php\RuleSupport::resolvedName($useUse->name);
                if (null === $resolved) {
                    $resolved = $useUse->name->toString();
                }
                $mapping = $analysisContext->findByOldSymbol($resolved);
                if (null === $mapping) {
                    $updatedUses[] = \Refactorlah\PhpAdapter\Php\RuleSupport::text($context, $useUse);
                    continue;
                }

                if ($this->shouldRemoveImport($useUse, $mapping, $effectiveNamespace)) {
                    $changed = true;
                    continue;
                }

                $updatedUses[] = $this->renderUseUse($useUse, $mapping->newSymbol);
                $changed = true;
            }

            if (!$changed) {
                continue;
            }

            $replacement = $this->statementReplacement($context, $useStatement, $updatedUses);
            if (null !== $replacement) {
                $replacements[] = $replacement;
            }
        }

        return $replacements;
    }

    private function shouldRemoveImport(UseUse $useUse, \Refactorlah\PhpAdapter\Php\SymbolMapping $mapping, string $effectiveNamespace): bool
    {
        if ('' === $effectiveNamespace || null !== $useUse->alias) {
            return false;
        }

        return $mapping->newNamespace === $effectiveNamespace
            && $useUse->name->getLast() === $mapping->shortName;
    }

    /** @param list<string> $updatedUses */
    private function statementReplacement(PhpFileContext $context, Use_ $useStatement, array $updatedUses): ?Replacement
    {
        $replacement = [] === $updatedUses
            ? ''
            : $this->renderUseStatement($useStatement, $updatedUses);

        $start = $useStatement->getStartFilePos();
        $end = $useStatement->getEndFilePos();
        if (!is_int($start) || !is_int($end) || $start < 0 || $end < $start) {
            return null;
        }

        $endExclusive = $end + 1;
        if ('' === $replacement) {
            while ($endExclusive < mb_strlen($context->content)) {
                $char = $context->content[$endExclusive];
                if ("\n" !== $char && "\r" !== $char) {
                    break;
                }
                $endExclusive++;
            }
        }

        return new Replacement(
            file: $context->path,
            start: $start,
            end: $endExclusive,
            replacement: $replacement,
            reason: 'php-use-statement',
            rule: $this->name(),
        );
    }

    /** @param list<string> $updatedUses */
    private function renderUseStatement(Use_ $useStatement, array $updatedUses): string
    {
        $prefix = match ($useStatement->type) {
            Use_::TYPE_FUNCTION => 'use function ',
            Use_::TYPE_CONSTANT => 'use const ',
            default => 'use ',
        };

        return $prefix . implode(', ', $updatedUses) . ';';
    }

    private function renderUseUse(UseUse $useUse, string $symbol): string
    {
        $rendered = $symbol;
        if (null !== $useUse->alias) {
            $rendered .= ' as ' . $useUse->alias->toString();
        }

        return $rendered;
    }
}
