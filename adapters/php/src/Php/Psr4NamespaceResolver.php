<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use Refactorlah\PhpAdapter\Composer\Psr4Map;

use function array_filter;
use function array_pop;
use function array_values;
use function count;
use function explode;
use function implode;
use function mb_rtrim;
use function mb_strlen;
use function mb_substr;
use function mb_trim;
use function str_ends_with;
use function str_replace;
use function str_starts_with;

final class Psr4NamespaceResolver
{
    public function deriveSymbol(Psr4Map $map, string $relativePath): ?ResolvedSymbol
    {
        if (!str_ends_with($relativePath, '.php')) {
            return null;
        }

        $normalized = str_replace('\\', '/', $relativePath);
        $bestNamespace = null;
        $bestBasePath = null;

        foreach ($map->all() as $namespace => $paths) {
            foreach ($paths as $path) {
                $prefix = mb_trim($path, '/');
                if ('.' !== $prefix && !str_starts_with($normalized, $prefix . '/')) {
                    continue;
                }
                if ('.' === $prefix || str_starts_with($normalized, $prefix . '/')) {
                    if (null === $bestBasePath || mb_strlen($prefix) > mb_strlen($bestBasePath)) {
                        $bestNamespace = $namespace;
                        $bestBasePath = $prefix;
                    }
                }
            }
        }

        if (null === $bestNamespace) {
            return null;
        }
        $basePath = $bestBasePath;
        if (null === $basePath) {
            return null;
        }

        $relative = '.' === $basePath
            ? $normalized
            : mb_substr($normalized, mb_strlen($basePath) + 1);

        $withoutExtension = mb_substr($relative, 0, -4);
        $parts = array_values(array_filter(explode('/', $withoutExtension), static fn(string $part): bool => '' !== $part));
        if ([] === $parts) {
            return null;
        }

        $shortName = $parts[count($parts) - 1];
        array_pop($parts);
        $namespace = mb_rtrim($bestNamespace, '\\');
        if ([] !== $parts) {
            $namespace .= '\\' . implode('\\', $parts);
        }

        return new ResolvedSymbol(
            symbol: $namespace . '\\' . $shortName,
            namespace: $namespace,
            shortName: $shortName,
        );
    }
}
