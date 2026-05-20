<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

final class SymfonyTemplateAttributeReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/#\[\s*Template\(\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
