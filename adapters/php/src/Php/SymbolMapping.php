<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

final class SymbolMapping
{
    public function __construct(
        public readonly string $kind,
        public readonly string $oldPath,
        public readonly string $newPath,
        public readonly string $oldSymbol,
        public readonly string $newSymbol,
        public readonly string $oldNamespace,
        public readonly string $newNamespace,
        public readonly string $shortName,
    ) {}

    public function toArray(): array
    {
        return [
            'kind' => $this->kind,
            'oldPath' => $this->oldPath,
            'newPath' => $this->newPath,
            'oldSymbol' => $this->oldSymbol,
            'newSymbol' => $this->newSymbol,
            'oldNamespace' => $this->oldNamespace,
            'newNamespace' => $this->newNamespace,
            'shortName' => $this->shortName,
        ];
    }
}
