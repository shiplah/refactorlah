<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Warning\Warning;

use function array_map;

final class Response implements \JsonSerializable
{
    /**
     * @param list<array<string,mixed>> $symbolMappings
     * @param list<array<string,mixed>> $pathMappings
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
