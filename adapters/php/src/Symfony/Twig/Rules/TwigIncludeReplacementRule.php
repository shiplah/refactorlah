<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig\Rules;

use function preg_quote;

final class TwigIncludeReplacementRule extends \Refactorlah\PhpAdapter\Symfony\Twig\Rules\AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/{%\s*include\s+(' . preg_quote($quotedReference, '/') . ')/',
            '/\{\{\s*include\(\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
