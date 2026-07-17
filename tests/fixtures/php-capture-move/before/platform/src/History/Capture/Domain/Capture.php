<?php
namespace App\History\Capture\Domain;

final readonly class Capture
{
    public function __construct(
        public int $capturedAt,
        public string $captureKey,
    ) {}
}
