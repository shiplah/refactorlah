<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use function preg_quote;

final class TwigFromReplacementRule extends \Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*from\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
