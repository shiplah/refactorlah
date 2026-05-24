<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig;

final class TwigPathRoot
{
    public function __construct(
        public readonly string $path,
        public readonly ?string $namespace = null,
    ) {}
}
