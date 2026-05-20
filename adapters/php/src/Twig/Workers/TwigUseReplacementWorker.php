<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

final class TwigUseReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*use\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
