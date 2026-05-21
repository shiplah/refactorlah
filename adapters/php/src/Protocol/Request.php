<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use function array_values;
use function is_array;

/**
 * @phpstan-type RequestMove array{oldPath:string,newPath:string,tracked:bool}
 * @phpstan-type RequestOptions array{includePhp:bool,includeTwig:bool}
 * @phpstan-type RequestPayload array{
 *   protocolVersion:int,
 *   projectRoot:string,
 *   oldPath:string,
 *   newPath:string,
 *   dryRun:bool,
 *   moves:list<RequestMove>,
 *   options:RequestOptions
 * }
 */

final class Request
{
    /** @param list<RequestMove> $moves */
    public function __construct(
        public readonly int $protocolVersion,
        public readonly string $projectRoot,
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly bool $dryRun,
        public readonly array $moves,
        public readonly bool $includePhp,
        public readonly bool $includeTwig,
    ) {}

    /** @param array<string,mixed> $data */
    public static function fromArray(array $data): self
    {
        $options = self::normalizeOptions($data['options'] ?? null);

        return new self(
            protocolVersion: (int) ($data['protocolVersion'] ?? 0),
            projectRoot: (string) ($data['projectRoot'] ?? '.'),
            oldPath: (string) ($data['oldPath'] ?? ''),
            newPath: (string) ($data['newPath'] ?? ''),
            dryRun: (bool) ($data['dryRun'] ?? true),
            moves: self::normalizeMoves($data['moves'] ?? null),
            includePhp: $options['includePhp'],
            includeTwig: $options['includeTwig'],
        );
    }

    /**
     * @param mixed $moves
     * @return list<RequestMove>
     */
    private static function normalizeMoves(mixed $moves): array
    {
        if (!is_array($moves)) {
            return [];
        }

        $normalized = [];
        foreach ($moves as $move) {
            if (!is_array($move)) {
                continue;
            }

            $normalized[] = [
                'oldPath' => (string) ($move['oldPath'] ?? ''),
                'newPath' => (string) ($move['newPath'] ?? ''),
                'tracked' => (bool) ($move['tracked'] ?? false),
            ];
        }

        return array_values($normalized);
    }

    /**
     * @param mixed $options
     * @return RequestOptions
     */
    private static function normalizeOptions(mixed $options): array
    {
        if (!is_array($options)) {
            return [
                'includePhp' => false,
                'includeTwig' => false,
            ];
        }

        return [
            'includePhp' => (bool) ($options['includePhp'] ?? false),
            'includeTwig' => (bool) ($options['includeTwig'] ?? false),
        ];
    }
}
