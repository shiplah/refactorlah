<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Files;

final class TextFileDetector
{
    public function isText(string $path): bool
    {
        $content = @file_get_contents($path);
        if ($content === false) {
            return false;
        }

        return !str_contains(substr($content, 0, 8000), "\0");
    }
}
