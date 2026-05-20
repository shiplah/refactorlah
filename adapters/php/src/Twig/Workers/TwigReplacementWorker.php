<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Workers;

use Refactorlah\PhpAdapter\Replacement\Replacement;

interface TwigReplacementWorker
{
    public function name(): string;

    /**
     * @param array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string} $mapping
     * @return list<Replacement>
     */
    public function collect(string $file, string $content, array $mapping): array;
}
