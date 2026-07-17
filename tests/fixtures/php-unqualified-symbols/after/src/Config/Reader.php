<?php

namespace App\Config;

use const App\Shared\DEFAULT_LIMIT;
use function App\Shared\build_label;

final class Reader
{
    public function label(string $value): string
    {
        return build_label($value) . DEFAULT_LIMIT;
    }
}
