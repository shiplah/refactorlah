<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

use function preg_quote;

final class TwigExtendsReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*extends\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
