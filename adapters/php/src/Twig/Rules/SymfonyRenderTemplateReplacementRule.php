<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use function preg_quote;

final class SymfonyRenderTemplateReplacementRule extends \Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/->render(?:View)?\(\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
