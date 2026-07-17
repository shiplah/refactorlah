<?php
namespace App\Consumer;

use App\Module\Record\Domain\Record;

final readonly class RecordMapper
{
    #[Map]
    public function map(?object $source): object
    {
        $source ?? throw new \LogicException('Source data is required.');

        return new PairRecords(
            old: new Record(
                createdAt: 1_779_194_233,
                recordKey: $source?->olderRecordKey,
            ),
            new: new Record(
                createdAt: 1_779_448_907,
                recordKey: $source?->newerRecordKey,
            ),
        );
    }
}
