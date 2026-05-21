<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Warning;

/**
 * @phpstan-type WarningArray array{
 *   message:string,
 *   file?:string,
 *   line?:int
 * }
 */
final class Warning
{
    public function __construct(
        public readonly string $message,
        public readonly string $file = '',
        public readonly int $line = 0,
    ) {}

    /** @return WarningArray */
    public function toArray(): array
    {
        $data = ['message' => $this->message];
        if ('' !== $this->file) {
            $data['file'] = $this->file;
        }
        if ($this->line > 0) {
            $data['line'] = $this->line;
        }

        return $data;
    }
}
