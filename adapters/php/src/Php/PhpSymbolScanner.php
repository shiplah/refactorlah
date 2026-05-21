<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node;
use PhpParser\Node\Stmt;
use PhpParser\ParserFactory;
use Refactorlah\PhpAdapter\Composer\Psr4Map;
use Refactorlah\PhpAdapter\Warning\Warning;

use function basename;
use function count;
use function file_get_contents;
use function str_ends_with;

final class PhpSymbolScanner
{
    public function __construct(
        private readonly Psr4NamespaceResolver $resolver,
    ) {}

    /**
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     * @return array{0:list<SymbolMapping>,1:list<Warning>}
     */
    public function scan(string $projectRoot, Psr4Map $map, array $moves): array
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $mappings = [];
        $warnings = [];

        foreach ($moves as $move) {
            $oldPath = $move['oldPath'];
            $newPath = $move['newPath'];
            if (!str_ends_with($oldPath, '.php')) {
                continue;
            }

            $oldResolved = $this->resolver->deriveSymbol($map, $oldPath);
            $newResolved = $this->resolver->deriveSymbol($map, $newPath);
            if (null === $oldResolved || null === $newResolved) {
                $warnings[] = new Warning(
                    message: 'Moved PHP file is outside known PSR-4 roots; symbol mapping skipped.',
                    file: $oldPath,
                );
                continue;
            }

            $content = (string) file_get_contents($projectRoot . '/' . $oldPath);
            $ast = $parser->parse($content);
            if (null === $ast) {
                $warnings[] = new Warning(message: 'PHP file could not be parsed; symbol mapping skipped.', file: $oldPath);
                continue;
            }

            $symbols = $this->findTopLevelSymbols($ast);
            $shortName = basename($oldPath, '.php');
            $chosen = $this->chooseSymbol($symbols, $shortName);
            if (null === $chosen) {
                $warnings[] = new Warning(
                    message: 'Multiple or ambiguous top-level symbols detected; symbol mapping skipped.',
                    file: $oldPath,
                );
                continue;
            }

            $name = $chosen->name?->toString();
            if ($name !== $oldResolved->shortName) {
                $warnings[] = new Warning(
                    message: 'Top-level symbol does not match deterministic PSR-4 filename; symbol mapping skipped.',
                    file: $oldPath,
                );
                continue;
            }

            $mappings[] = new SymbolMapping(
                kind: $this->nodeKind($chosen),
                oldPath: $oldPath,
                newPath: $newPath,
                oldSymbol: $oldResolved->symbol,
                newSymbol: $newResolved->symbol,
                oldNamespace: $oldResolved->namespace,
                newNamespace: $newResolved->namespace,
                shortName: $oldResolved->shortName,
            );
        }

        return [$mappings, $warnings];
    }

    /**
     * @param list<Node> $ast
     * @return list<Stmt\Class_|Stmt\Interface_|Stmt\Trait_|Stmt\Enum_>
     */
    private function findTopLevelSymbols(array $ast): array
    {
        $symbols = [];
        foreach ($ast as $node) {
            if ($node instanceof Stmt\Namespace_) {
                foreach ($node->stmts as $stmt) {
                    if ($this->isPrimarySymbol($stmt)) {
                        $symbols[] = $stmt;
                    }
                }
                continue;
            }

            if ($this->isPrimarySymbol($node)) {
                $symbols[] = $node;
            }
        }

        return $symbols;
    }

    /** @param list<Stmt\Class_|Stmt\Interface_|Stmt\Trait_|Stmt\Enum_> $symbols */
    private function chooseSymbol(array $symbols, string $shortName): Stmt\Class_|Stmt\Interface_|Stmt\Trait_|Stmt\Enum_|null
    {
        if (1 === count($symbols)) {
            return $symbols[0];
        }

        foreach ($symbols as $symbol) {
            if ($symbol->name?->toString() === $shortName) {
                return $symbol;
            }
        }

        return null;
    }

    private function isPrimarySymbol(Node $node): bool
    {
        return $node instanceof Stmt\Class_
            || $node instanceof Stmt\Interface_
            || $node instanceof Stmt\Trait_
            || $node instanceof Stmt\Enum_;
    }

    private function nodeKind(Node $node): string
    {
        return match (true) {
            $node instanceof Stmt\Interface_ => 'interface',
            $node instanceof Stmt\Trait_ => 'trait',
            $node instanceof Stmt\Enum_ => 'enum',
            default => 'class',
        };
    }
}
