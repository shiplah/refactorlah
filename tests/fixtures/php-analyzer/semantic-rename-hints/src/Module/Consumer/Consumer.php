<?php
namespace App\Module\Consumer;

final class Consumer
{
    public function __construct(private iterable $oldItemHandlers) {}
    public function tag(): string { return 'app.old_item_handler'; }
}
