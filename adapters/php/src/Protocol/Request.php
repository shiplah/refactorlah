<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use function array_key_exists;
use function is_array;
use function is_bool;
use function is_int;
use function is_string;

/**
 * @phpstan-type RequestOptions array{includePhp:bool,includeTwig:bool,scanInclude:list<string>,scanExclude:list<string>}
 * @phpstan-type RequestPayload array{
 *   protocolVersion:int,
 *   projectRoot:string,
 *   oldPath:string,
 *   newPath:string,
 *   dryRun:bool,
 *   moves:list<array{oldPath:string,newPath:string,tracked:bool}>,
 *   options:RequestOptions
 * }
 */

final class Request
{
    public function __construct(
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly MoveCollection $moves,
        public readonly bool $includePhp,
        public readonly bool $includeTwig,
        /** @var list<string> */
        public readonly array $scanInclude,
        /** @var list<string> */
        public readonly array $scanExclude,
    ) {}

    /** @param array<string,mixed> $data */
    public static function fromArray(array $data): self
    {
        self::validatePayload($data);
        $options = self::normalizeOptions($data['options'] ?? null);

        return new self(
            oldPath: self::mixedString($data['oldPath'] ?? ''),
            newPath: self::mixedString($data['newPath'] ?? ''),
            moves: MoveCollection::fromMixed($data['moves'] ?? null),
            includePhp: $options['includePhp'],
            includeTwig: $options['includeTwig'],
            scanInclude: $options['scanInclude'],
            scanExclude: $options['scanExclude'],
        );
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
                'scanInclude' => [],
                'scanExclude' => [],
            ];
        }

        return [
            'includePhp' => (bool) ($options['includePhp'] ?? false),
            'includeTwig' => (bool) ($options['includeTwig'] ?? false),
            'scanInclude' => self::stringList($options['scanInclude'] ?? []),
            'scanExclude' => self::stringList($options['scanExclude'] ?? []),
        ];
    }

    /** @return list<string> */
    private static function stringList(mixed $value): array
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

    private static function mixedInt(mixed $value): int
    {
        return is_int($value) ? $value : 0;
    }

    private static function mixedString(mixed $value): string
    {
        return is_string($value) ? $value : '';
    }

    /** @param array<string,mixed> $data */
    private static function validatePayload(array $data): void
    {
        if (1 !== self::mixedInt($data['protocolVersion'] ?? null)) {
            throw new \RuntimeException('adapter request must use protocolVersion 1');
        }

        if ('.' !== self::mixedString($data['projectRoot'] ?? null)) {
            throw new \RuntimeException('adapter request must use projectRoot "."');
        }

        if ('' === self::mixedString($data['oldPath'] ?? null) || '' === self::mixedString($data['newPath'] ?? null)) {
            throw new \RuntimeException('adapter request must include oldPath and newPath');
        }

        if (!array_key_exists('dryRun', $data) || !is_bool($data['dryRun'])) {
            throw new \RuntimeException('adapter request must include dryRun');
        }

        if (MoveCollection::fromMixed($data['moves'] ?? null)->isEmpty()) {
            throw new \RuntimeException('adapter request must include at least one move');
        }
    }
}
