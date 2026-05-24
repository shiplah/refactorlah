<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Protocol;

use function is_array;
use function is_string;

/**
 * @phpstan-type RequestOptionsArray array{
 *   includePhp:bool,
 *   includeTwig:bool,
 *   scanInclude:list<string>,
 *   scanExclude:list<string>
 * }
 */
final class RequestOptions
{
    /**
     * @param list<string> $scanInclude
     * @param list<string> $scanExclude
     */
    public function __construct(
        public readonly bool $includePhp,
        public readonly bool $includeTwig,
        public readonly array $scanInclude,
        public readonly array $scanExclude,
    ) {}

    public static function fromMixed(mixed $options): self
    {
        if (!is_array($options)) {
            return new self(
                includePhp: false,
                includeTwig: false,
                scanInclude: [],
                scanExclude: [],
            );
        }

        return new self(
            includePhp: (bool) ($options['includePhp'] ?? false),
            includeTwig: (bool) ($options['includeTwig'] ?? false),
            scanInclude: self::stringList($options['scanInclude'] ?? []),
            scanExclude: self::stringList($options['scanExclude'] ?? []),
        );
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
}
