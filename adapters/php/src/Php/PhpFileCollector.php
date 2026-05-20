<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use Refactorlah\PhpAdapter\Files\FileCollector;

final class PhpFileCollector
{
    public function __construct(private readonly FileCollector $collector)
    {
    }

    /**
     * @return list<string>
     */
    public function collect(string $projectRoot): array
    {
        return $this->collector->collect($projectRoot, ['php']);
    }
}
