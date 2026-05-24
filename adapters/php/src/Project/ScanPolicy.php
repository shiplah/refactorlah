<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Project;

use function preg_match;
use function preg_quote;
use function array_filter;
use function array_values;
use function str_replace;

final class ScanPolicy
{
    /**
     * @param list<string> $include
     * @param list<string> $exclude
     */
    public function __construct(
        private readonly array $include,
        private readonly array $exclude,
    ) {}

    public function allows(string $path): bool
    {
        foreach ($this->include as $pattern) {
            if ($this->matches($pattern, $path)) {
                return true;
            }
        }

        foreach ($this->exclude as $pattern) {
            if ($this->matches($pattern, $path)) {
                return false;
            }
        }

        return true;
    }

    /**
     * @param list<string> $files
     * @return list<string>
     */
    public function filter(array $files): array
    {
        return array_values(array_filter($files, fn(string $file): bool => $this->allows($file)));
    }

    private function matches(string $pattern, string $path): bool
    {
        $quoted = preg_quote($pattern, '/');
        $quoted = str_replace('\*\*', '.*', $quoted);
        $quoted = str_replace('\*', '[^/]*', $quoted);
        $quoted = str_replace('\?', '[^/]', $quoted);

        return 1 === preg_match('/^' . $quoted . '$/', $path);
    }
}
