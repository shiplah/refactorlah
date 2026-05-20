<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Composer;

final class Psr4Map
{
    /**
     * @param array<string,list<string>> $mappings
     */
    public function __construct(private readonly array $mappings)
    {
    }

    /**
     * @return array<string,list<string>>
     */
    public function all(): array
    {
        return $this->mappings;
    }
}
