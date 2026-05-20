<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Replacement;

final class ReplacementCollection
{
    /** @var list<Replacement> */
    private array $items = [];

    public function add(Replacement ...$replacements): void
    {
        foreach ($replacements as $replacement) {
            $this->items[] = $replacement;
        }
    }

    /**
     * @return list<Replacement>
     */
    public function all(): array
    {
        return $this->items;
    }
}
