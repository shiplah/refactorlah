<?php
namespace App\Example;

use App\Old\Thing;

final class Consumer
{
    public function build(): \App\Old\Thing
    {
        return new \App\Old\Thing();
    }
}
