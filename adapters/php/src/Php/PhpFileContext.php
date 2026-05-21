<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node;

final class PhpFileContext
{
    /** @param list<Node> $ast */
    public function __construct(
        public readonly string $path,
        public readonly string $content,
        public readonly array $ast,
    ) {}
}
