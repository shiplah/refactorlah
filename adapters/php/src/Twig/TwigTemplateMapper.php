<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

final class TwigTemplateMapper
{
    /**
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     * @return list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}>
     */
    public function deriveMappings(array $moves, TwigPathConfiguration $configuration): array
    {
        $mappings = [];
        foreach ($moves as $move) {
            $oldPath = $move['oldPath'];
            $newPath = $move['newPath'];
            if (!str_ends_with($oldPath, '.twig') || !str_ends_with($newPath, '.twig')) {
                continue;
            }

            $oldReference = $this->referenceForPath($oldPath, $configuration);
            $newReference = $this->referenceForPath($newPath, $configuration);
            if ($oldReference === null || $newReference === null) {
                continue;
            }

            $mappings[] = [
                'kind' => 'twig-template',
                'oldPath' => $oldPath,
                'newPath' => $newPath,
                'oldReference' => $oldReference,
                'newReference' => $newReference,
            ];
        }

        return $mappings;
    }

    private function referenceForPath(string $path, TwigPathConfiguration $configuration): ?string
    {
        $bestRoot = null;
        foreach ($configuration->roots as $root) {
            if ($path === $root->path || str_starts_with($path, $root->path . '/')) {
                if ($bestRoot === null || strlen($root->path) > strlen($bestRoot->path)) {
                    $bestRoot = $root;
                }
            }
        }

        if ($bestRoot === null) {
            return null;
        }

        $relative = ltrim(substr($path, strlen($bestRoot->path)), '/');
        if ($bestRoot->namespace === null || $bestRoot->namespace === '') {
            return $relative;
        }

        return '@' . $bestRoot->namespace . '/' . $relative;
    }
}
