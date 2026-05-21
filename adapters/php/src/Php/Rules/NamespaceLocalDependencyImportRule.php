<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Expr\Instanceof_;
use PhpParser\Node\Expr\New_;
use PhpParser\Node\Expr\StaticCall;
use PhpParser\Node\Expr\StaticPropertyFetch;
use PhpParser\Node\Name;
use PhpParser\Node\Stmt\Catch_;
use PhpParser\Node\Stmt\Class_;
use PhpParser\Node\Stmt\Enum_;
use PhpParser\Node\Stmt\GroupUse;
use PhpParser\Node\Stmt\Interface_;
use PhpParser\Node\Stmt\Namespace_;
use PhpParser\Node\Stmt\TraitUse;
use PhpParser\Node\Stmt\UseUse;
use PhpParser\Node\Stmt\Use_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\RuleSupport;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function array_map;
use function array_values;
use function basename;
use function implode;
use function in_array;
use function is_int;
use function mb_strrpos;
use function mb_substr;
use function pathinfo;
use function sort;
use function str_contains;

final class NamespaceLocalDependencyImportRule implements ReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $declaredNamespace = RuleSupport::declaredNamespace($context);
        if ('' === $declaredNamespace) {
            return [];
        }

        $effectiveNamespace = RuleSupport::effectiveNamespace($context, $analysisContext);

        $finder = new NodeFinder();
        /** @var list<Name> $names */
        $names = $finder->findInstanceOf($context->ast, Name::class);
        $existingImports = $this->existingImports($context);
        $plannedImports = [];
        $replacements = [];

        foreach ($names as $name) {
            if (!$this->shouldInspect($name)) {
                continue;
            }

            $resolved = RuleSupport::resolvedName($name);
            if (null === $resolved || !$this->belongsToDeclaredNamespace($resolved, $declaredNamespace)) {
                continue;
            }

            $desiredSymbol = $analysisContext->findByOldSymbol($resolved)?->newSymbol ?? $resolved;
            if ($this->namespaceOf($desiredSymbol) === $effectiveNamespace) {
                continue;
            }

            $shortName = $this->shortName($desiredSymbol);
            if (RuleSupport::importsSymbol($context, $desiredSymbol, $shortName)) {
                continue;
            }

            if (($existingImports[$shortName] ?? null) !== null && $existingImports[$shortName] !== $desiredSymbol) {
                $replacement = RuleSupport::createReplacement(
                    $context,
                    $name,
                    '\\' . $desiredSymbol,
                    'php-namespace-local-reference',
                    $this->name(),
                );
                if (null !== $replacement) {
                    $replacements[] = $replacement;
                }
                continue;
            }

            if (($plannedImports[$shortName] ?? null) !== null && $plannedImports[$shortName] !== $desiredSymbol) {
                $replacement = RuleSupport::createReplacement(
                    $context,
                    $name,
                    '\\' . $desiredSymbol,
                    'php-namespace-local-reference',
                    $this->name(),
                );
                if (null !== $replacement) {
                    $replacements[] = $replacement;
                }
                continue;
            }

            if ($shortName === $this->currentFileShortName($context->path)) {
                $replacement = RuleSupport::createReplacement(
                    $context,
                    $name,
                    '\\' . $desiredSymbol,
                    'php-namespace-local-reference',
                    $this->name(),
                );
                if (null !== $replacement) {
                    $replacements[] = $replacement;
                }
                continue;
            }

            $plannedImports[$shortName] = $desiredSymbol;
        }

        if ([] === $plannedImports) {
            return $replacements;
        }

        $insertion = $this->buildNamespaceInsertion($context, array_values($plannedImports));
        if (null !== $insertion) {
            $replacements[] = $insertion;
        }

        return $replacements;
    }

    private function shouldInspect(Name $name): bool
    {
        $original = $name->getAttribute('originalName');
        if (!$original instanceof Name || !$original->isUnqualified()) {
            return false;
        }

        $parent = $name->getAttribute('parent');
        if ($parent instanceof UseUse || $parent instanceof Use_ || $parent instanceof GroupUse || $parent instanceof Namespace_) {
            return false;
        }

        if (RuleSupport::isTypeReference($name)) {
            return true;
        }

        return match (true) {
            $parent instanceof New_ => $parent->class === $name,
            $parent instanceof StaticCall => $parent->class === $name,
            $parent instanceof StaticPropertyFetch => $parent->class === $name,
            $parent instanceof ClassConstFetch => $parent->class === $name,
            $parent instanceof Instanceof_ => $parent->class === $name,
            $parent instanceof Catch_ => true,
            $parent instanceof TraitUse => true,
            $parent instanceof Class_ => $parent->extends === $name || in_array($name, $parent->implements, true),
            $parent instanceof Interface_ => in_array($name, $parent->extends, true),
            $parent instanceof Enum_ => in_array($name, $parent->implements, true),
            default => false,
        };
    }

    private function belongsToDeclaredNamespace(string $resolved, string $declaredNamespace): bool
    {
        return str_contains($resolved, '\\')
            && $this->namespaceOf($resolved) === $declaredNamespace;
    }

    /** @return array<string, string> */
    private function existingImports(PhpFileContext $context): array
    {
        $finder = new NodeFinder();
        /** @var list<Use_> $useStatements */
        $useStatements = $finder->findInstanceOf($context->ast, Use_::class);

        $imports = [];
        foreach ($useStatements as $useStatement) {
            if ($useStatement instanceof GroupUse) {
                continue;
            }

            if (Use_::TYPE_NORMAL !== $useStatement->type) {
                continue;
            }

            foreach ($useStatement->uses as $useUse) {
                $resolved = RuleSupport::resolvedName($useUse->name) ?? $useUse->name->toString();
                $alias = $useUse->alias?->toString() ?? $useUse->name->getLast();
                $imports[$alias] = $resolved;
            }
        }

        return $imports;
    }

    /** @param list<string> $symbols */
    private function buildNamespaceInsertion(PhpFileContext $context, array $symbols): ?Replacement
    {
        sort($symbols);

        $finder = new NodeFinder();
        /** @var Namespace_|null $namespace */
        $namespace = $finder->findFirstInstanceOf($context->ast, Namespace_::class);
        if (!$namespace instanceof Namespace_) {
            return null;
        }

        $offset = $namespace->getEndFilePos();
        if (!is_int($offset) || $offset < 0) {
            return null;
        }

        return new Replacement(
            file: $context->path,
            start: $offset + 1,
            end: $offset + 1,
            replacement: "\n\n" . $this->renderImports($symbols),
            reason: 'php-namespace-local-import',
            rule: $this->name(),
        );
    }

    /** @param list<string> $symbols */
    private function renderImports(array $symbols): string
    {
        return implode("\n", array_map(
            static fn(string $symbol): string => 'use ' . $symbol . ';',
            $symbols,
        ));
    }

    private function namespaceOf(string $symbol): string
    {
        $index = mb_strrpos($symbol, '\\');
        if (false === $index) {
            return '';
        }

        return mb_substr($symbol, 0, $index);
    }

    private function shortName(string $symbol): string
    {
        $index = mb_strrpos($symbol, '\\');
        if (false === $index) {
            return $symbol;
        }

        return mb_substr($symbol, $index + 1);
    }

    private function currentFileShortName(string $path): string
    {
        $filename = basename($path);
        $shortName = pathinfo($filename, PATHINFO_FILENAME);

        return '' === $shortName ? $filename : $shortName;
    }
}
