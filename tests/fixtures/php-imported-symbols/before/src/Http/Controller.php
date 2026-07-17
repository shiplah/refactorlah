<?php

namespace App\Http;

use const App\Config\DEFAULT_LIMIT;
use function App\Config\build_label;

final class Controller
{
    public function label(string $value): string
    {
        return build_label($value) . DEFAULT_LIMIT;
    }
}
