<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use function preg_quote;

final class YamlTwigComponentTemplateDirectoryReplacementRule extends AbstractTwigStringReplacementRule
{
    protected function patterns(string $quotedReference): array
    {
        return ['/\btemplate_directory:\s*(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
