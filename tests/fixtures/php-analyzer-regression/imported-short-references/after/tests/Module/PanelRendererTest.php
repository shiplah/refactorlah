<?php

declare(strict_types=1);

namespace App\Tests\Module;

use App\Module\Ui\Renderer\PanelRenderableRenderer;

$renderer = new PanelRenderableRenderer();
$matches = $renderer instanceof PanelRenderableRenderer;
