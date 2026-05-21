<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

final class ResolvedSymbol
{
    public function __construct(
        public readonly string $symbol,
        public readonly string $namespace,
        public readonly string $shortName,
    ) {}
}
