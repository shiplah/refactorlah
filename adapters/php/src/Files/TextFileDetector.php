<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Files;

use function file_get_contents;
use function mb_substr;
use function str_contains;

final class TextFileDetector
{
    public function isText(string $path): bool
    {
        $content = @file_get_contents($path);
        if (false === $content) {
            return false;
        }

        return !str_contains(mb_substr($content, 0, 8000), "\0");
    }
}
