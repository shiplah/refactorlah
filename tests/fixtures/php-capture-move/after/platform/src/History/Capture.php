<?php
namespace App\History;

final readonly class Capture
{
    public function __construct(
        public int $capturedAt,
        public string $captureKey,
    ) {}
}
