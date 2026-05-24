<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

final class Move
{
    public function __construct(
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly bool $tracked,
    ) {}
}
