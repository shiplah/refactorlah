<?php
namespace App\History\ComparisonDocument\Application;

use App\History\Capture\Domain\Capture;
use App\History\Capture\Domain\CaptureCollection;

final readonly class DocumentPageDataMapper
{
    public function map(?object $artifacts): CaptureCollection
    {
        $artifacts ?? throw new \LogicException('Rendered artifacts are required.');

        new ComparisonCaptures(
            old: new Capture(
                capturedAt: 1_779_194_233,
                captureKey: $artifacts?->olderCaptureKey,
            ),
            new: new Capture(
                capturedAt: 1_779_448_907,
                captureKey: $artifacts?->newerCaptureKey,
            ),
        );

        return new CaptureCollection();
    }
}
