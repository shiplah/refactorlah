<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use Refactorlah\PhpAdapter\Project\ProjectContext;

use function array_map;
use function array_values;
use function count;

/**
 * @phpstan-import-type PathMappingArray from \Refactorlah\PhpAdapter\Config\PathMapping
 * @implements \IteratorAggregate<int, PathMapping>
 */
final class PathMappingCollection implements \Countable, \IteratorAggregate
{
    /** @param list<PathMapping> $mappings */
    public function __construct(
        private readonly array $mappings = [],
    ) {}

    public static function empty(): self
    {
        return new self();
    }

    public function count(): int
    {
        return count($this->mappings);
    }

    public function isEmpty(): bool
    {
        return [] === $this->mappings;
    }

    /** @return \Traversable<int, PathMapping> */
    public function getIterator(): \Traversable
    {
        yield from $this->mappings;
    }

    public function with(PathMapping $mapping): self
    {
        return new self([...$this->mappings, $mapping]);
    }

    public function withUnique(PathMapping $mapping): self
    {
        foreach ($this->mappings as $existing) {
            if ($existing->identity() === $mapping->identity()) {
                return $this;
            }
        }

        return $this->with($mapping);
    }

    public function merge(self $other): self
    {
        $merged = $this;
        foreach ($other as $mapping) {
            $merged = $merged->withUnique($mapping);
        }

        return $merged;
    }

    public function toProjectRelative(ProjectContext $context): self
    {
        $mappings = [];
        foreach ($this->mappings as $mapping) {
            $mappings[] = $mapping->toProjectRelative($context);
        }

        return new self($mappings);
    }

    public function containsOldReference(string $content): bool
    {
        foreach ($this->mappings as $mapping) {
            if ($mapping->oldReferenceOccursIn($content)) {
                return true;
            }
        }

        return false;
    }

    /** @return list<string> */
    public function warningIndicators(): array
    {
        $indicators = [];
        foreach ($this->mappings as $mapping) {
            foreach ($mapping->warningIndicators() as $indicator) {
                $indicators[$indicator] = $indicator;
            }
        }

        return array_values($indicators);
    }

    /** @return list<PathMapping> */
    public function values(): array
    {
        return $this->mappings;
    }

    /** @return list<PathMappingArray> */
    public function toArray(): array
    {
        return array_map(static fn(PathMapping $mapping): array => $mapping->toArray(), $this->mappings);
    }
}
