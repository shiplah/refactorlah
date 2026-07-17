<?php

declare(strict_types=1);

namespace App\Module\Ui;

use App\Module\Ui\Block\PanelBlockRenderer;

final class HtmlRenderer
{
    private ?PanelBlockRenderer $renderer = null;

    public function render(PanelBlockRenderer $renderer): PanelBlockRenderer
    {
        $this->renderer = $renderer;

        if (!$renderer instanceof PanelBlockRenderer) {
            return new PanelBlockRenderer();
        }

        return PanelBlockRenderer::make();
    }
}
