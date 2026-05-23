<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

use function file_get_contents;
use function is_array;
use function is_file;
use function is_string;
use function json_decode;
use function mb_strlen;
use function mb_substr;
use function str_starts_with;

final class RefactorlahConfigReader
{
    public function readForContext(string $projectRoot, ProjectContext $context): RefactorlahConfig
    {
        $rootPatterns = $this->readPatterns($projectRoot);
        if ('.' === $context->subRoot) {
            return new RefactorlahConfig(
                include: $rootPatterns['include'],
                exclude: $rootPatterns['exclude'],
            );
        }

        $subRootPatterns = $this->readPatterns($context->absoluteRoot);

        return new RefactorlahConfig(
            include: [
                ...$this->patternsForSubRoot($rootPatterns['include'], $context->subRoot),
                ...$subRootPatterns['include'],
            ],
            exclude: [
                ...$this->patternsForSubRoot($rootPatterns['exclude'], $context->subRoot),
                ...$subRootPatterns['exclude'],
            ],
        );
    }

    /** @return array{include:list<string>,exclude:list<string>} */
    private function readPatterns(string $projectRoot): array
    {
        $path = $projectRoot . '/.refactorlah.json';
        if (!is_file($path)) {
            return ['include' => [], 'exclude' => []];
        }

        $decoded = json_decode((string) file_get_contents($path), true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($decoded)) {
            return ['include' => [], 'exclude' => []];
        }

        return [
            'include' => $this->stringList($decoded['include'] ?? []),
            'exclude' => $this->stringList($decoded['exclude'] ?? []),
        ];
    }

    /**
     * @param list<string> $patterns
     * @return list<string>
     */
    private function patternsForSubRoot(array $patterns, string $subRoot): array
    {
        $prefix = $subRoot . '/';
        $relative = [];

        foreach ($patterns as $pattern) {
            if (str_starts_with($pattern, $prefix)) {
                $relative[] = mb_substr($pattern, mb_strlen($prefix));
                continue;
            }

            if (str_starts_with($pattern, '**/')) {
                $relative[] = $pattern;
            }
        }

        return $relative;
    }

    /** @return list<string> */
    private function stringList(mixed $value): array
    {
        if (!is_array($value)) {
            return [];
        }

        $strings = [];
        foreach ($value as $item) {
            if (is_string($item) && '' !== $item) {
                $strings[] = $item;
            }
        }

        return $strings;
    }
}
