<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

use Refactorlah\PhpAdapter\Files\FileCollector;

final class TwigReferenceScanner
{
    public function __construct(private readonly FileCollector $collector)
    {
    }

    /**
     * @return list<string>
     */
    public function collectTwigFiles(string $projectRoot): array
    {
        return $this->collector->collect($projectRoot, ['twig']);
    }

    /**
     * @return list<string>
     */
    public function collectConfigFiles(string $projectRoot): array
    {
        return $this->collector->collect($projectRoot, ['php', 'yaml', 'yml']);
    }
}
