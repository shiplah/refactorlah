<?php

namespace App\Config;

final class Reader
{
    public function label(string $value): string
    {
        return build_label($value) . DEFAULT_LIMIT;
    }
}
