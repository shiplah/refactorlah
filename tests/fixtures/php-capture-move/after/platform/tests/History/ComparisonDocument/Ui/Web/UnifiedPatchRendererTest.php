<?php
namespace App\Tests\History\ComparisonDocument\Ui\Web;

use App\History\Capture;

final class UnifiedPatchRendererTest
{
    #[Test]
    public function itRenders(): void
    {
        new Capture(1_779_194_233, '2026-05-19-1237');
    }
}
