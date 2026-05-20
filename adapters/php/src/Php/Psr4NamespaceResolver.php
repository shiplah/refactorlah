<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use Refactorlah\PhpAdapter\Composer\Psr4Map;

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
                $prefix = trim($path, '/');
                if ($prefix !== '.' && !str_starts_with($normalized, $prefix . '/')) {
                    continue;
                }
                if ($prefix === '.' || str_starts_with($normalized, $prefix . '/')) {
                    if ($bestBasePath === null || strlen($prefix) > strlen($bestBasePath)) {
                        $bestNamespace = $namespace;
                        $bestBasePath = $prefix;
                    }
                }
            }
        }

        if ($bestNamespace === null) {
            return null;
        }

        $relative = $bestBasePath === '.'
            ? $normalized
            : substr($normalized, strlen($bestBasePath) + 1);

        $withoutExtension = substr($relative, 0, -4);
        $parts = array_values(array_filter(explode('/', $withoutExtension), static fn (string $part): bool => $part !== ''));
        if ($parts === []) {
            return null;
        }

        $shortName = array_pop($parts);
        $namespace = rtrim($bestNamespace, '\\');
        if ($parts !== []) {
            $namespace .= '\\' . implode('\\', $parts);
        }

        return new ResolvedSymbol(
            symbol: $namespace . '\\' . $shortName,
            namespace: $namespace,
            shortName: $shortName,
        );
    }
}
