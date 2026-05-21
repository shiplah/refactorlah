<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

use function array_intersect;
use function array_merge;
use function array_unique;
use function array_values;
use function dirname;
use function is_file;
use function mb_rtrim;
use function mb_strlen;
use function str_contains;
use function str_ends_with;
use function str_replace;
use function usort;

final class ProjectContextResolver
{
    /** @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves */
    public function resolve(string $projectRoot, array $moves): ProjectContext
    {
        $candidateRoots = null;

        foreach ($moves as $move) {
            $moveRoots = array_unique(array_merge(
                $this->composerAncestors($projectRoot, $move['oldPath']),
                $this->composerAncestors($projectRoot, $move['newPath']),
            ));

            if ([] === $moveRoots) {
                continue;
            }

            $candidateRoots = null === $candidateRoots
                ? $moveRoots
                : array_values(array_intersect($candidateRoots, $moveRoots));
        }

        if (null === $candidateRoots || [] === $candidateRoots) {
            if (is_file($projectRoot . '/composer.json')) {
                return new ProjectContext('.', $projectRoot);
            }

            throw new \RuntimeException('composer.json is required for PHP adapter analysis');
        }

        usort($candidateRoots, static fn(string $left, string $right): int => mb_strlen($right) <=> mb_strlen($left));
        $subRoot = $candidateRoots[0];
        $absoluteRoot = '.' === $subRoot ? $projectRoot : $projectRoot . '/' . $subRoot;

        return new ProjectContext($subRoot, $absoluteRoot);
    }

    /** @return list<string> */
    private function composerAncestors(string $projectRoot, string $path): array
    {
        $normalized = str_replace('\\', '/', $path);
        $directory = str_contains($normalized, '.') && !str_ends_with($normalized, '/')
            ? dirname($normalized)
            : mb_rtrim($normalized, '/');

        $candidates = [];
        while ('.' !== $directory && '' !== $directory) {
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
