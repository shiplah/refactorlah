<?php

declare(strict_types=1);

namespace App\Module\Record\Application;

use App\Shared\Input\InputRecord;
use App\Module\Record\Domain\MappedValue;

final readonly class RecordMapper
{
    public function map(InputRecord $record): MappedValue
    {
        return new MappedValue();
    }
}
