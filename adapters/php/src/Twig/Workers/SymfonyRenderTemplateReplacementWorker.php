<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

use function preg_quote;

final class SymfonyRenderTemplateReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/->render(?:View)?\(\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
