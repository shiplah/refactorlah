<?php

declare(strict_types=1);

namespace App\Module\Ui\Renderer;

final class PanelRenderableRenderer
{
    public static function make(): self
    {
        return new self();
    }
}
