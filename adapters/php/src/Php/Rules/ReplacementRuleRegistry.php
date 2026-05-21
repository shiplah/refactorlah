<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

use function array_merge;

final class ReplacementRuleRegistry
{
    /** @var list<ReplacementRule> */
    private array $rules;

    public function __construct(ReplacementRule ...$rules)
    {
        $this->rules = $rules;
    }

    /** @return list<\Refactorlah\PhpAdapter\Replacement\Replacement> */
    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $replacements = [];
        foreach ($this->rules as $rule) {
            $replacements = array_merge($replacements, $rule->collect($context, $analysisContext));
        }

        return $replacements;
    }
}
