<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig\Rules;

use Refactorlah\PhpAdapter\Config\PathMapping;
use Refactorlah\PhpAdapter\Replacement\Replacement;

use function mb_strlen;
use function mb_strtolower;
use function preg_match_all;

abstract class AbstractTwigStringReplacementRule
{
    /** @return list<string> */
    abstract protected function patterns(string $quotedReference): array;

    /** @return list<Replacement> */
    public function collect(string $file, string $content, PathMapping $mapping): array
    {
        $replacements = [];
        foreach ($mapping->quotedOldReferences() as $quotedReference) {
            foreach ($this->patterns($quotedReference) as $pattern) {
                if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
                    continue;
                }

                foreach ($matches[1] as [$matched, $offset]) {
                    $replacements[] = new Replacement(
                        file: $file,
                        start: $offset,
                        end: $offset + mb_strlen($matched),
                        replacement: $mapping->replacementForQuotedReference($quotedReference),
                        reason: mb_strtolower((new \ReflectionClass($this))->getShortName()),
                        rule: static::class,
                    );
                }
            }
        }

        return $replacements;
    }
}
