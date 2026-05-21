<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig\Rules;

use Refactorlah\PhpAdapter\Replacement\Replacement;

use function mb_strlen;
use function mb_strtolower;
use function preg_match_all;

abstract class AbstractTwigStringReplacementRule implements \Refactorlah\PhpAdapter\Twig\Rules\TwigReplacementRule
{
    /** @return list<string> */
    abstract protected function patterns(string $quotedReference): array;

    /**
     * @param array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string} $mapping
     * @return list<Replacement>
     */
    public function collect(string $file, string $content, array $mapping): array
    {
        $replacements = [];
        foreach (["'" . $mapping['oldReference'] . "'", '"' . $mapping['oldReference'] . '"'] as $quotedReference) {
            foreach ($this->patterns($quotedReference) as $pattern) {
                if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
                    continue;
                }

                foreach ($matches[1] as [$matched, $offset]) {
                    $replacementValue = $quotedReference[0] . $mapping['newReference'] . $quotedReference[0];
                    $replacements[] = new Replacement(
                        file: $file,
                        start: $offset,
                        end: $offset + mb_strlen($matched),
                        replacement: $replacementValue,
                        reason: mb_strtolower((new \ReflectionClass($this))->getShortName()),
                        rule: $this->name(),
                    );
                }
            }
        }

        return $replacements;
    }

    public function name(): string
    {
        return static::class;
    }
}
