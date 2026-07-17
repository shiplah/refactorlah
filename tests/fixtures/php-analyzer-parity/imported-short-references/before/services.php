<?php

use App\Module\Ui\Block\PanelBlockRenderer;

return static function ($services): void {
    $services->instanceof(PanelBlockRenderer::class);
    $services->set(PanelBlockRenderer::class);
};
