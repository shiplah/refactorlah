<?php

use App\Module\Ui\Renderer\PanelRenderableRenderer;

return static function ($services): void {
    $services->instanceof(PanelRenderableRenderer::class);
    $services->set(PanelRenderableRenderer::class);
};
