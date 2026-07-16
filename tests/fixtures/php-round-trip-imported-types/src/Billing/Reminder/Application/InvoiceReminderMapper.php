<?php

declare(strict_types=1);

namespace App\Billing\Reminder\Application;

use App\Schema\Model\InvoiceReminder;
use App\Billing\Reminder\Domain\ReminderMessage;

final readonly class InvoiceReminderMapper
{
    public function map(InvoiceReminder $notice): ReminderMessage
    {
        return new ReminderMessage();
    }
}
