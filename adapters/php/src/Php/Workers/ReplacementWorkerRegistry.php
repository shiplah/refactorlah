<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class ReplacementWorkerRegistry
{
    /** @var list<ReplacementWorker> */
    private array $workers;

    public function __construct(ReplacementWorker ...$workers)
    {
        $this->workers = $workers;
    }

    /**
     * @return list<\Refactorlah\PhpAdapter\Replacement\Replacement>
     */
    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $replacements = [];
        foreach ($this->workers as $worker) {
            $replacements = array_merge($replacements, $worker->collect($context, $analysisContext));
        }

        return $replacements;
    }
}
