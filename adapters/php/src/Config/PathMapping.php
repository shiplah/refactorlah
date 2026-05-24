<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use Refactorlah\PhpAdapter\Project\ProjectContext;

/**
 * @phpstan-type PathMappingArray array{
 *   kind:string,
 *   oldPath:string,
 *   newPath:string,
 *   oldReference:string,
 *   newReference:string
 * }
 */
final class PathMapping
{
    public function __construct(
        public readonly string $kind,
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly string $oldReference,
        public readonly string $newReference,
    ) {}

    public function toProjectRelative(ProjectContext $context): self
    {
        return new self(
            kind: $this->kind,
            oldPath: $context->toProjectRelative($this->oldPath),
            newPath: $context->toProjectRelative($this->newPath),
            oldReference: $this->oldReference,
            newReference: $this->newReference,
        );
    }

    /** @return PathMappingArray */
    public function toArray(): array
    {
        return [
            'kind' => $this->kind,
            'oldPath' => $this->oldPath,
            'newPath' => $this->newPath,
            'oldReference' => $this->oldReference,
            'newReference' => $this->newReference,
        ];
    }
}
