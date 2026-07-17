<?php

declare(strict_types=1);

namespace App\Module\Ui\Renderer;

use App\Module\Application\DirectiveKind;

final class PanelRenderer
{
    public function kind(): DirectiveKind
    {
        return DirectiveKind::Panel;
    }
}
