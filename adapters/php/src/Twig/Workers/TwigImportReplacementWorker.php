<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

use function preg_quote;

final class TwigImportReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/{%\s*import\s+(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
