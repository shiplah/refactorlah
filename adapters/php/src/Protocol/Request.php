<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

final class Request
{
    /**
     * @param list<array{oldPath:string,newPath:string,tracked:bool}> $moves
     */
    public function __construct(
        public readonly int $protocolVersion,
        public readonly string $projectRoot,
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly bool $dryRun,
        public readonly array $moves,
        public readonly bool $includePhp,
        public readonly bool $includeTwig,
    ) {
    }

    public static function fromArray(array $data): self
    {
        return new self(
            protocolVersion: (int) ($data['protocolVersion'] ?? 0),
            projectRoot: (string) ($data['projectRoot'] ?? '.'),
            oldPath: (string) ($data['oldPath'] ?? ''),
            newPath: (string) ($data['newPath'] ?? ''),
            dryRun: (bool) ($data['dryRun'] ?? true),
            moves: array_values($data['moves'] ?? []),
            includePhp: (bool) (($data['options']['includePhp'] ?? false)),
            includeTwig: (bool) (($data['options']['includeTwig'] ?? false)),
        );
    }
}
