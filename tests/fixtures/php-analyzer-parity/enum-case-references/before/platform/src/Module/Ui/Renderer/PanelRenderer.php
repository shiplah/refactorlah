<?php

declare(strict_types=1);

namespace App\Module\Ui\Renderer;

use App\Module\Application\ElementKind;

final class PanelRenderer
{
    public function kind(): ElementKind
    {
        return ElementKind::Panel;
    }
}
