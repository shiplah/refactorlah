<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

final class YamlTwigTemplateReplacementWorker extends AbstractTwigStringReplacementWorker
{
    protected function patterns(string $quotedReference): array
    {
        return ['/\btemplate:\s*(' . preg_quote($quotedReference, '/') . ')/'];
    }
}
