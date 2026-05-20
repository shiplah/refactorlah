<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Warning;

final class WarningCollection
{
    /** @var list<Warning> */
    private array $items = [];

    public function add(Warning ...$warnings): void
    {
        foreach ($warnings as $warning) {
            $this->items[] = $warning;
        }
    }

    /**
     * @return list<Warning>
     */
    public function all(): array
    {
        return $this->items;
    }
}
