<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

final class TwigEmbedReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*embed\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
