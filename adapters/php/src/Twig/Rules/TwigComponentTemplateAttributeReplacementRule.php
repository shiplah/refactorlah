<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use function preg_quote;

final class TwigComponentTemplateAttributeReplacementRule extends AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return [
            '/#\[[^\]]*\bAsTwigComponent\b[^\]]*\btemplate\s*:\s*(' . preg_quote($quotedReference, '/') . ')/',
        ];
    }
}
