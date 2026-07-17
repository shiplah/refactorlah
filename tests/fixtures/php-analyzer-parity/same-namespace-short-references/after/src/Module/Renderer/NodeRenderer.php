<?php

declare(strict_types=1);

namespace App\Module\Renderer;

final class NodeRenderer
{
    /** @param iterable<DirectiveRenderer> $renderers */
    public function __construct(private iterable $renderers) {}

    public function renderer(?DirectiveRenderer $renderer): DirectiveRenderer
    {
        if (!$renderer instanceof DirectiveRenderer) {
            return new DirectiveRenderer();
        }

        return DirectiveRenderer::make();
    }
}
