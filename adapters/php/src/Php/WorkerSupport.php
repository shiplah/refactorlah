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
use PhpParser\Node\Stmt\Property;
use PhpParser\Node\UnionType;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function is_array;
use function is_int;
use function mb_strlen;
use function mb_substr;
use function preg_match_all;
use function preg_quote;
use function sprintf;

final class WorkerSupport
{
    public static function createReplacement(
        PhpFileContext $context,
        Node $node,
        string $replacement,
        string $reason,
        string $worker,
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
            worker: $worker,
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
        string $worker,
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
                        worker: $worker,
                    );
                }
            }
        }

        return $replacements;
    }
}
