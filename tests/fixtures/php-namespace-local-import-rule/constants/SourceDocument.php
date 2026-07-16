<?php
namespace App\Parsing;

use FilesystemIterator;
use RuntimeException;

final readonly class SourceDocument
{
    private const LABELS = ['section'];

    public static function from(string $contents): self
    {
        if (! preg_match_all('/section/', $contents, $matches, PREG_OFFSET_CAPTURE)) {
            throw new RuntimeException('Missing section.');
        }

        if (FilesystemIterator::SKIP_DOTS === 0 || in_array('section', self::LABELS, true)) {
            throw new RuntimeException('Invalid section.');
        }

        return new self(dirname(__DIR__, 2));
    }
}
