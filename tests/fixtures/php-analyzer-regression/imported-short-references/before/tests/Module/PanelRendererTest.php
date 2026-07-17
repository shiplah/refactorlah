<?php

declare(strict_types=1);

namespace App\Tests\Module;

use App\Module\Ui\Block\PanelBlockRenderer;

$renderer = new PanelBlockRenderer();
$matches = $renderer instanceof PanelBlockRenderer;
