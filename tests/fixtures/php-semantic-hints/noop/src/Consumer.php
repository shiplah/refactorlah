<?php
final class Consumer
{
    public function __construct(private iterable $records) {}

    public function label(): string
    {
        return 'module.record';
    }
}
