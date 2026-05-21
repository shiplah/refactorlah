<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

use function preg_quote;

final class TwigIncludeReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/{%\s*include\s+(' . preg_quote($quotedReference, '/') . ')/',
            '/\{\{\s*include\(\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
