<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Config;

use function mb_rtrim;

final class PathMappingFactory
{
    /** @return list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> */
    public function fromMove(string $oldPath, string $newPath): array
    {
        $oldReference = mb_rtrim($oldPath, '/') . '/';
        $newReference = mb_rtrim($newPath, '/') . '/';
        if ('/' === $oldReference || '/' === $newReference || $oldReference === $newReference) {
            return [];
        }

        return [[
            'kind' => 'project-path-directory',
            'oldPath' => mb_rtrim($oldPath, '/'),
            'newPath' => mb_rtrim($newPath, '/'),
            'oldReference' => $oldReference,
            'newReference' => $newReference,
        ]];
    }
}
