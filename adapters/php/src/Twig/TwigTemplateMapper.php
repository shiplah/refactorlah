<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

final class TwigTemplateMapper
{
    /**
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     * @return list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}>
     */
    public function deriveMappings(array $moves): array
    {
        $mappings = [];
        foreach ($moves as $move) {
            $oldPath = $move['oldPath'];
            $newPath = $move['newPath'];
            if (!str_ends_with($oldPath, '.twig') || !str_starts_with($oldPath, 'templates/')) {
                continue;
            }
            if (!str_starts_with($newPath, 'templates/')) {
                continue;
            }

            $mappings[] = [
                'kind' => 'twig-template',
                'oldPath' => $oldPath,
                'newPath' => $newPath,
                'oldReference' => substr($oldPath, strlen('templates/')),
                'newReference' => substr($newPath, strlen('templates/')),
            ];
        }

        return $mappings;
    }
}
