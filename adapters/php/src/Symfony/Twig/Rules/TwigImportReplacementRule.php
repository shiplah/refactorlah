<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig\Rules;

use function preg_quote;

final class TwigImportReplacementRule extends \Refactorlah\PhpAdapter\Symfony\Twig\Rules\AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*import\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
