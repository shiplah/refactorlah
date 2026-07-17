<?php

declare(strict_types=1);

namespace App\Module\Ui;

use App\Module\Ui\Renderer\PanelRenderableRenderer;

final class HtmlRenderer
{
    private ?PanelRenderableRenderer $renderer = null;

    public function render(PanelRenderableRenderer $renderer): PanelRenderableRenderer
    {
        $this->renderer = $renderer;

        if (!$renderer instanceof PanelRenderableRenderer) {
            return new PanelRenderableRenderer();
        }

        return PanelRenderableRenderer::make();
    }
}
