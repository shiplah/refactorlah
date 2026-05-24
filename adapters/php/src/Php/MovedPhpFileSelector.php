<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use Refactorlah\PhpAdapter\Protocol\MoveCollection;

use function str_ends_with;

final class MovedPhpFileSelector
{
    /** @return list<string> */
    public function oldPaths(MoveCollection $moves): array
    {
        $paths = [];
        foreach ($moves as $move) {
            if (str_ends_with($move->oldPath, '.php')) {
                $paths[] = $move->oldPath;
            }
        }

        return $paths;
    }
}
