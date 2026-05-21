<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Warning\Warning;

use function array_map;

/**
 * @phpstan-import-type SymbolMappingArray from \Refactorlah\PhpAdapter\Php\SymbolMapping
 * @phpstan-import-type ReplacementArray from \Refactorlah\PhpAdapter\Replacement\Replacement
 * @phpstan-import-type WarningArray from \Refactorlah\PhpAdapter\Warning\Warning
 * @phpstan-type PathMappingArray array{
 *   kind:string,
 *   oldPath:string,
 *   newPath:string,
 *   oldReference:string,
 *   newReference:string
 * }
 * @phpstan-type ResponsePayload array{
 *   protocolVersion:int,
 *   adapter:string,
 *   symbolMappings:list<SymbolMappingArray>,
 *   pathMappings:list<PathMappingArray>,
 *   replacements:list<ReplacementArray>,
 *   warnings:list<WarningArray>,
 *   errors:list<string>
 * }
 */
final class Response implements \JsonSerializable
{
    /**
     * @param list<SymbolMappingArray> $symbolMappings
     * @param list<PathMappingArray> $pathMappings
     * @param list<Replacement> $replacements
     * @param list<Warning> $warnings
     * @param list<string> $errors
     */
    public function __construct(
        private readonly array $symbolMappings,
        private readonly array $pathMappings,
        private readonly array $replacements,
        private readonly array $warnings,
        private readonly array $errors,
    ) {}

    /** @return ResponsePayload */
    public function jsonSerialize(): array
    {
        return [
            'protocolVersion' => 1,
            'adapter' => 'php',
            'symbolMappings' => $this->symbolMappings,
            'pathMappings' => $this->pathMappings,
            'replacements' => array_map(static fn(Replacement $replacement) => $replacement->toArray(), $this->replacements),
            'warnings' => array_map(static fn(Warning $warning) => $warning->toArray(), $this->warnings),
            'errors' => $this->errors,
        ];
    }
}
