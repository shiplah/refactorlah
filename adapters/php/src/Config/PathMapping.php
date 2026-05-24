<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use Refactorlah\PhpAdapter\Project\ProjectContext;

use function basename;
use function str_contains;

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

    public function identity(): string
    {
        return $this->kind . "\0" . $this->oldReference . "\0" . $this->newReference;
    }

    public function oldReferenceOccursIn(string $content): bool
    {
        return str_contains($content, $this->oldReference);
    }

    /** @return list<string> */
    public function warningIndicators(): array
    {
        return [$this->oldReference, basename($this->oldReference)];
    }

    /** @return list<string> */
    public function quotedOldReferences(): array
    {
        return ["'" . $this->oldReference . "'", '"' . $this->oldReference . '"'];
    }

    public function replacementForQuotedReference(string $quotedReference): string
    {
        $quote = $quotedReference[0] ?? "'";

        return $quote . $this->newReference . $quote;
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
