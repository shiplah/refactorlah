<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

use RuntimeException;

final class ProjectContextResolver
{
    /**
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     */
    public function resolve(string $projectRoot, array $moves): ProjectContext
    {
        $candidateRoots = null;

        foreach ($moves as $move) {
            $moveRoots = array_unique(array_merge(
                $this->composerAncestors($projectRoot, $move['oldPath']),
                $this->composerAncestors($projectRoot, $move['newPath']),
            ));

            if ($moveRoots === []) {
                continue;
            }

            $candidateRoots = $candidateRoots === null
                ? $moveRoots
                : array_values(array_intersect($candidateRoots, $moveRoots));
        }

        if ($candidateRoots === null || $candidateRoots === []) {
            if (is_file($projectRoot . '/composer.json')) {
                return new ProjectContext('.', $projectRoot);
            }

            throw new RuntimeException('composer.json is required for PHP adapter analysis');
        }

        usort($candidateRoots, static fn (string $left, string $right): int => strlen($right) <=> strlen($left));
        $subRoot = $candidateRoots[0];
        $absoluteRoot = $subRoot === '.' ? $projectRoot : $projectRoot . '/' . $subRoot;

        return new ProjectContext($subRoot, $absoluteRoot);
    }

    /**
     * @return list<string>
     */
    private function composerAncestors(string $projectRoot, string $path): array
    {
        $normalized = str_replace('\\', '/', $path);
        $directory = str_contains($normalized, '.') && !str_ends_with($normalized, '/')
            ? dirname($normalized)
            : rtrim($normalized, '/');

        $candidates = [];
        while ($directory !== '.' && $directory !== '') {
            if (is_file($projectRoot . '/' . $directory . '/composer.json')) {
                $candidates[] = $directory;
            }
            $next = dirname($directory);
            if ($next === $directory) {
                break;
            }
            $directory = $next;
        }

        if (is_file($projectRoot . '/composer.json')) {
            $candidates[] = '.';
        }

        return $candidates;
    }
}
