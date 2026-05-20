<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

final class TwigFromReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*from\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
