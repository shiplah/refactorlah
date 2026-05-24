<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use Refactorlah\PhpAdapter\Project\ProjectContext;

use function count;
use function is_array;
use function is_string;

/**
 * @implements \IteratorAggregate<int, Move>
 */
final class MoveCollection implements \Countable, \IteratorAggregate
{
    /** @param list<Move> $moves */
    public function __construct(
        private readonly array $moves,
    ) {}

    /** @param mixed $moves */
    public static function fromMixed(mixed $moves): self
    {
        if (!is_array($moves)) {
            return new self([]);
        }

        $normalised = [];
        foreach ($moves as $move) {
            if (!is_array($move)) {
                continue;
            }

            $normalised[] = new Move(
                oldPath: self::mixedString($move['oldPath'] ?? ''),
                newPath: self::mixedString($move['newPath'] ?? ''),
                tracked: (bool) ($move['tracked'] ?? false),
            );
        }

        return new self($normalised);
    }

    public function count(): int
    {
        return count($this->moves);
    }

    public function isEmpty(): bool
    {
        return [] === $this->moves;
    }

    /** @return \Traversable<int, Move> */
    public function getIterator(): \Traversable
    {
        yield from $this->moves;
    }

    public function toSubRootRelative(ProjectContext $context): self
    {
        $moves = [];
        foreach ($this->moves as $move) {
            $moves[] = new Move(
                oldPath: $context->toSubRootRelative($move->oldPath),
                newPath: $context->toSubRootRelative($move->newPath),
                tracked: $move->tracked,
            );
        }

        return new self($moves);
    }

    private static function mixedString(mixed $value): string
    {
        return is_string($value) ? $value : '';
    }
}
