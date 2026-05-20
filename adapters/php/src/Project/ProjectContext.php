<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

final class ProjectContext
{
    public function __construct(
        public readonly string $subRoot,
        public readonly string $absoluteRoot,
    ) {
    }

    public function toSubRootRelative(string $path): string
    {
        if ($this->subRoot === '.') {
            return $path;
        }

        if (str_starts_with($path, $this->subRoot . '/')) {
            return substr($path, strlen($this->subRoot) + 1);
        }

        return $path;
    }

    public function toProjectRelative(string $path): string
    {
        if ($this->subRoot === '.') {
            return $path;
        }

        return $this->subRoot . '/' . ltrim($path, '/');
    }
}
