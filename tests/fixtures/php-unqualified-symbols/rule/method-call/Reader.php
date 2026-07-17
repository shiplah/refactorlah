<?php

namespace App\Config;

final class Reader
{
    public function label(object $formatter, string $value): string
    {
        return $formatter->build_label($value);
    }
}
