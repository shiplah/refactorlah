<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig;

final class TwigPathConfiguration
{
    /** @param list<TwigPathRoot> $roots */
    public function __construct(
        public readonly array $roots,
    ) {}
}
