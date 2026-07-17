<?php
namespace App\History\Capture\Domain;

use App\Shared\Support\Collection;
use App\History\Capture;

use function array_reverse;
use function usort;

final readonly class CaptureCollection extends Collection
{
    public function previous(Capture $capture): ?Capture
    {
        return $capture;
    }
}
