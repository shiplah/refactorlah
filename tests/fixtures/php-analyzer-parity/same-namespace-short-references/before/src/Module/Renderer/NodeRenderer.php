<?php

declare(strict_types=1);

namespace App\Module\Renderer;

final class NodeRenderer
{
    /** @param iterable<ComponentRenderer> $renderers */
    public function __construct(private iterable $renderers) {}

    public function renderer(?ComponentRenderer $renderer): ComponentRenderer
    {
        if (!$renderer instanceof ComponentRenderer) {
            return new ComponentRenderer();
        }

        return ComponentRenderer::make();
    }
}
