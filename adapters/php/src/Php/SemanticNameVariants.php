<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use function array_keys;
use function array_unique;
use function array_values;
use function implode;
use function mb_strrpos;
use function mb_strtolower;
use function mb_substr;
use function preg_match_all;

final class SemanticNameVariants
{
    /** @return list<string> */
    public function needles(SymbolMapping $mapping): array
    {
        $variableHints = $this->variableHints($mapping);
        $literalHints = $this->literalHints($mapping);

        return array_values(array_unique([
            ...array_keys($variableHints),
            ...array_keys($literalHints),
        ]));
    }

    /** @return array<string,string> */
    public function variableHints(SymbolMapping $mapping): array
    {
        $oldLowerCamel = $this->lowerFirst($this->shortName($mapping->oldSymbol));
        $newLowerCamel = $this->lowerFirst($this->shortName($mapping->newSymbol));

        return [
            $oldLowerCamel => $newLowerCamel,
            $oldLowerCamel . 's' => $newLowerCamel . 's',
        ];
    }

    /** @return array<string,string> */
    public function literalHints(SymbolMapping $mapping): array
    {
        $oldShortName = $this->shortName($mapping->oldSymbol);
        $newShortName = $this->shortName($mapping->newSymbol);
        $oldSnake = $this->toDelimited($oldShortName, '_');
        $newSnake = $this->toDelimited($newShortName, '_');
        $oldKebab = $this->toDelimited($oldShortName, '-');
        $newKebab = $this->toDelimited($newShortName, '-');

        return [
            $this->lowerFirst($oldShortName) => $this->lowerFirst($newShortName),
            $this->lowerFirst($oldShortName) . 's' => $this->lowerFirst($newShortName) . 's',
            $oldSnake => $newSnake,
            $oldSnake . 's' => $newSnake . 's',
            $oldKebab => $newKebab,
            $oldKebab . 's' => $newKebab . 's',
        ];
    }

    public function shortName(string $symbol): string
    {
        $index = mb_strrpos($symbol, '\\');
        if (false === $index) {
            return $symbol;
        }

        return mb_substr($symbol, $index + 1);
    }

    private function toDelimited(string $name, string $delimiter): string
    {
        if (!preg_match_all('/[A-Z]+(?=[A-Z][a-z]|$)|[A-Z]?[a-z0-9]+/', $name, $matches)) {
            return mb_strtolower($name);
        }

        return mb_strtolower(implode($delimiter, $matches[0]));
    }

    private function lowerFirst(string $value): string
    {
        if ('' === $value) {
            return '';
        }

        return mb_strtolower(mb_substr($value, 0, 1)) . mb_substr($value, 1);
    }
}
