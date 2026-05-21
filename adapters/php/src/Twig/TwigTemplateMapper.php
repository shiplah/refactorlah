<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

use function mb_ltrim;
use function mb_strlen;
use function mb_substr;
use function str_ends_with;
use function str_starts_with;

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
            if (null === $oldReference || null === $newReference) {
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
                if (null === $bestRoot || mb_strlen($root->path) > mb_strlen($bestRoot->path)) {
                    $bestRoot = $root;
                }
            }
        }

        if (null === $bestRoot) {
            return null;
        }

        $relative = mb_ltrim(mb_substr($path, mb_strlen($bestRoot->path)), '/');
        if (null === $bestRoot->namespace || '' === $bestRoot->namespace) {
            return $relative;
        }

        return '@' . $bestRoot->namespace . '/' . $relative;
    }
}
