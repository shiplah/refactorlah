<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use function mb_rtrim;

final class PathMappingFactory
{
    public function fromMove(string $oldPath, string $newPath): PathMappingCollection
    {
        $oldReference = mb_rtrim($oldPath, '/') . '/';
        $newReference = mb_rtrim($newPath, '/') . '/';
        if ('/' === $oldReference || '/' === $newReference || $oldReference === $newReference) {
            return PathMappingCollection::empty();
        }

        return PathMappingCollection::empty()->with(new PathMapping(
            kind: 'project-path-directory',
            oldPath: mb_rtrim($oldPath, '/'),
            newPath: mb_rtrim($newPath, '/'),
            oldReference: $oldReference,
            newReference: $newReference,
        ));
    }
}
