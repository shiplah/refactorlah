<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node;
use PhpParser\Node\Expr\ArrowFunction;
use PhpParser\Node\Expr\Closure;
use PhpParser\Node\IntersectionType;
use PhpParser\Node\Name;
use PhpParser\Node\NullableType;
use PhpParser\Node\Param;
use PhpParser\Node\Stmt\ClassMethod;
use PhpParser\Node\Stmt\Function_;
use PhpParser\Node\Stmt\GroupUse;
use PhpParser\Node\Stmt\Namespace_;
use PhpParser\Node\Stmt\Property;
use PhpParser\Node\Stmt\UseUse;
use PhpParser\Node\Stmt\Use_;
use PhpParser\Node\UnionType;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function is_array;
use function is_int;
use function mb_strlen;
use function mb_substr;
use function preg_match_all;
use function preg_quote;
use function sprintf;

final class RuleSupport
{
    public static function createReplacement(
        PhpFileContext $context,
        Node $node,
        string $replacement,
        string $reason,
        string $rule,
    ): ?Replacement {
        $start = $node->getStartFilePos();
        $end = $node->getEndFilePos();
        if (!is_int($start) || !is_int($end) || $start < 0 || $end < $start) {
            return null;
        }

        return new Replacement(
            file: $context->path,
            start: $start,
            end: $end + 1,
            replacement: $replacement,
            reason: $reason,
            rule: $rule,
        );
    }

    public static function resolvedName(Name $name): ?string
    {
        $resolved = $name->getAttribute('resolvedName');
        if ($resolved instanceof Name) {
            return $resolved->toString();
        }

        $namespaced = $name->getAttribute('namespacedName');
        if ($namespaced instanceof Name) {
            return $namespaced->toString();
        }

        if ($name instanceof Name\FullyQualified) {
            return $name->toString();
        }

        return null;
    }

    public static function text(PhpFileContext $context, Node $node): string
    {
        $start = $node->getStartFilePos();
        $end = $node->getEndFilePos();
        if (!is_int($start) || !is_int($end) || $start < 0 || $end < $start) {
            return '';
        }

        return mb_substr($context->content, $start, $end - $start + 1);
    }

    public static function replacementName(PhpFileContext $context, Name $name, SymbolMapping $mapping): string
    {
        // Preserve the file's original reference style when that style remains valid
        // after the import/use rewrite. We only expand to an FQCN when short syntax
        // would become ambiguous or invalid after the move.
        $original = $name->getAttribute('originalName');
        if ($original instanceof Name) {
            if ($original instanceof Name\FullyQualified) {
                return '\\' . $mapping->newSymbol;
            }

            if ($original->isUnqualified() && self::importsSymbol($context, $mapping->oldSymbol, $original->toString())) {
                return $original->toString();
            }
        }

        if ($name instanceof Name\FullyQualified) {
            return '\\' . $mapping->newSymbol;
        }

        if ($name->isUnqualified() && self::importsSymbol($context, $mapping->oldSymbol, $name->toString())) {
            return $name->toString();
        }

        return '\\' . $mapping->newSymbol;
    }

    public static function inAttribute(Node $node): bool
    {
        $parent = $node->getAttribute('parent');
        while ($parent instanceof Node) {
            if ($parent instanceof Node\Attribute || $parent instanceof Node\AttributeGroup) {
                return true;
            }
            $parent = $parent->getAttribute('parent');
        }

        return false;
    }

    public static function isTypeReference(Name $name): bool
    {
        $current = $name;
        $parent = $name->getAttribute('parent');
        while ($parent instanceof NullableType || $parent instanceof UnionType || $parent instanceof IntersectionType) {
            $current = $parent;
            $parent = $parent->getAttribute('parent');
        }

        if ($parent instanceof Property && $parent->type === $current) {
            return true;
        }

        if ($parent instanceof Param && $parent->type === $current) {
            return true;
        }

        if (($parent instanceof ClassMethod || $parent instanceof Function_ || $parent instanceof Closure || $parent instanceof ArrowFunction)
            && $parent->getReturnType() === $current) {
            return true;
        }

        return false;
    }

    public static function importsSymbol(PhpFileContext $context, string $symbol, string $reference): bool
    {
        $finder = new NodeFinder();
        /** @var list<Use_> $useStatements */
        $useStatements = $finder->findInstanceOf($context->ast, Use_::class);

        foreach ($useStatements as $useStatement) {
            if ($useStatement instanceof GroupUse) {
                continue;
            }

            foreach ($useStatement->uses as $useUse) {
                if (!$useUse instanceof UseUse) {
                    continue;
                }

                $resolved = self::resolvedName($useUse->name);
                if (null === $resolved) {
                    $resolved = $useUse->name->toString();
                }
                if ($resolved !== $symbol) {
                    continue;
                }

                $alias = $useUse->alias?->toString() ?? $useUse->name->getLast();

                if ($alias === $reference) {
                    return true;
                }
            }
        }

        return false;
    }

    public static function effectiveNamespace(PhpFileContext $context, AnalysisContext $analysisContext): string
    {
        $mapping = $analysisContext->findByPath($context->path);
        if (null !== $mapping) {
            return $mapping->newNamespace;
        }

        $finder = new NodeFinder();
        /** @var Namespace_|null $namespace */
        $namespace = $finder->findFirstInstanceOf($context->ast, Namespace_::class);
        if ($namespace instanceof Namespace_ && null !== $namespace->name) {
            return $namespace->name->toString();
        }

        return '';
    }

    public static function attachParents(array $ast): void
    {
        foreach ($ast as $node) {
            self::attachParent($node, null);
        }
    }

    private static function attachParent(Node $node, ?Node $parent): void
    {
        $node->setAttribute('parent', $parent);
        foreach ($node->getSubNodeNames() as $name) {
            $child = $node->$name;
            if ($child instanceof Node) {
                self::attachParent($child, $node);
                continue;
            }

            if (is_array($child)) {
                foreach ($child as $nested) {
                    if ($nested instanceof Node) {
                        self::attachParent($nested, $node);
                    }
                }
            }
        }
    }

    public static function docblockTagReplacements(
        PhpFileContext $context,
        string $tag,
        AnalysisContext $analysisContext,
        string $reason,
        string $rule,
    ): array {
        $replacements = [];
        $pattern = sprintf('/@%s\b[^\n\r]*/', preg_quote($tag, '/'));
        if (!preg_match_all($pattern, $context->content, $matches, PREG_OFFSET_CAPTURE)) {
            return [];
        }

        foreach ($matches[0] as [$lineText, $lineOffset]) {
            foreach ($analysisContext->symbolMappings as $mapping) {
                $symbolPattern = '/(?<![A-Za-z0-9_\\\\])' . preg_quote($mapping->oldSymbol, '/') . '(?![A-Za-z0-9_\\\\])/';
                if (!preg_match_all($symbolPattern, $lineText, $symbolMatches, PREG_OFFSET_CAPTURE)) {
                    continue;
                }

                foreach ($symbolMatches[0] as [$matchedText, $matchOffset]) {
                    $replacements[] = new Replacement(
                        file: $context->path,
                        start: $lineOffset + $matchOffset,
                        end: $lineOffset + $matchOffset + mb_strlen($matchedText),
                        replacement: $mapping->newSymbol,
                        reason: $reason,
                        rule: $rule,
                    );
                }
            }
        }

        return $replacements;
    }
}
