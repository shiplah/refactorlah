<?php
namespace App\Config;

const DEFAULT_LIMIT = 10, SECOND_LIMIT = 20;

function build_label(string $value): string
{
    return $value;
}

final class LocalType
{
    public const CLASS_LIMIT = 30;
}
