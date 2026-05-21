<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Replacement;

/**
 * @phpstan-type ReplacementArray array{
 *   file:string,
 *   start:int,
 *   end:int,
 *   replacement:string,
 *   reason:string,
 *   rule:string
 * }
 */
final class Replacement
{
    public function __construct(
        public readonly string $file,
        public readonly int $start,
        public readonly int $end,
        public readonly string $replacement,
        public readonly string $reason,
        public readonly string $rule,
    ) {}

    /** @return ReplacementArray */
    public function toArray(): array
    {
        return [
            'file' => $this->file,
            'start' => $this->start,
            'end' => $this->end,
            'replacement' => $this->replacement,
            'reason' => $this->reason,
            'rule' => $this->rule,
        ];
    }
}
