<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use function preg_quote;

final class TwigUseReplacementRule extends \Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*use\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
