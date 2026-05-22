<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

use function file_get_contents;
use function is_array;
use function is_file;
use function is_string;
use function json_decode;

final class RefactorlahConfigReader
{
    public function read(string $projectRoot): RefactorlahConfig
    {
        $path = $projectRoot . '/.refactorlah.json';
        if (!is_file($path)) {
            return new RefactorlahConfig([], []);
        }

        $decoded = json_decode((string) file_get_contents($path), true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($decoded)) {
            return new RefactorlahConfig([], []);
        }

        return new RefactorlahConfig(
            include: $this->stringList($decoded['include'] ?? []),
            exclude: $this->stringList($decoded['exclude'] ?? []),
        );
    }

    /** @return list<string> */
    private function stringList(mixed $value): array
    {
        if (!is_array($value)) {
            return [];
        }

        $strings = [];
        foreach ($value as $item) {
            if (is_string($item) && '' !== $item) {
                $strings[] = $item;
            }
        }

        return $strings;
    }
}
