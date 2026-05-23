<?php

declare(strict_types=1);

test('analyze command emits valid protocol response for fixture project', function (): void
{
    $repoRoot = \dirname(__DIR__, 4);
    $fixtureRoot = $repoRoot . '/adapters/php/tests/fixtures/php-basic';
    $adapterBinary = $repoRoot . '/adapters/php/bin/refactorlah-php';
    $request = [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'app/Services/Billing',
        'newPath' => 'app/Domain/Billing',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'app/Services/Billing/InvoiceService.php',
            'newPath' => 'app/Domain/Billing/InvoiceService.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => true,
        ],
    ];

    $decoded = run_adapter($fixtureRoot, $request);
    assertSameValue(1, $decoded['protocolVersion']);
    assertSameValue('php', $decoded['adapter']);
    assertTrueValue(\count($decoded['symbolMappings']) >= 1, 'expected symbol mappings');
});

test('analyze command updates reordered namespace moves and dependent imports', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Billing/Domain/Archive', 0o777, true);
    \mkdir($root . '/src/Consumer', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Billing/Domain/Archive/InvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain\Archive;

        final class InvoiceLine {}
        PHP);
    \file_put_contents($root . '/src/Consumer/UsesInvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Consumer;

        use App\Billing\Domain\Archive\InvoiceLine;

        final class UsesInvoiceLine
        {
            public function make(): InvoiceLine
            {
                return new InvoiceLine();
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Billing/Domain/Archive',
        'newPath' => 'src/Billing/Archive/Domain',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Billing/Domain/Archive/InvoiceLine.php',
            'newPath' => 'src/Billing/Archive/Domain/InvoiceLine.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertSameValue('App\\Billing\\Archive\\Domain\\InvoiceLine', $decoded['symbolMappings'][0]['newSymbol']);
    assertTrueValue(
        has_replacement($decoded['replacements'], 'src/Billing/Domain/Archive/InvoiceLine.php', 'php-namespace-declaration', 'App\\Billing\\Archive\\Domain'),
        'expected moved file namespace replacement',
    );
    assertTrueValue(
        has_replacement($decoded['replacements'], 'src/Consumer/UsesInvoiceLine.php', 'php-use-statement', 'use App\\Billing\\Archive\\Domain\\InvoiceLine;'),
        'expected dependent use statement replacement',
    );
    assertTrueValue(
        has_replacement($decoded['replacements'], 'src/Consumer/UsesInvoiceLine.php', 'php-method-return-type', 'InvoiceLine'),
        'expected imported return type to stay short',
    );
});

test('analyze command updates moved test namespaces from autoload-dev psr4 roots', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/tests/Application/Billing/Document/ContentFix', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
        'autoload-dev' => [
            'psr-4' => [
                'App\\Tests\\' => 'tests/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/tests/Application/Billing/Document/ContentFix/ClaudeDocsDocumentationIndexFixerTest.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Application\Billing\Invoice\ContentFix;

        final class ClaudeDocsDocumentationIndexFixerTest {}
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'tests/Application/Billing/Document/ContentFix',
        'newPath' => 'tests/Billing/Archive/Detailed/Application/ContentFix',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'tests/Application/Billing/Document/ContentFix/ClaudeDocsDocumentationIndexFixerTest.php',
            'newPath' => 'tests/Billing/Archive/Detailed/Application/ContentFix/ClaudeDocsDocumentationIndexFixerTest.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertSameValue(
        'App\\Tests\\Billing\\Archive\\Detailed\\Application\\ContentFix\\ClaudeDocsDocumentationIndexFixerTest',
        $decoded['symbolMappings'][0]['newSymbol'],
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'tests/Application/Billing/Document/ContentFix/ClaudeDocsDocumentationIndexFixerTest.php',
            'php-namespace-declaration',
            'App\\Tests\\Billing\\Archive\\Detailed\\Application\\ContentFix',
        ),
        'expected moved test namespace replacement',
    );
});

test('analyze command updates moved file namespace inside nested composer roots', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceBatch {}
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceBatch.php',
            'php-namespace-declaration',
            'App\\Billing\\Archive\\Domain',
        ),
        'expected nested composer root namespace replacement',
    );
});

test('analyze command updates moved file class name when basename changes', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/app/Services/Billing', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'app/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    $original = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Services\Billing;

        final class InvoiceService {}
        PHP;
    \file_put_contents($root . '/app/Services/Billing/InvoiceService.php', $original);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Services/Billing/BillingService.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'app/Services/Billing/InvoiceService.php',
            'newPath' => 'app/Services/Billing/BillingService.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updated = apply_replacements_for_file($original, $decoded['replacements'], 'app/Services/Billing/InvoiceService.php');
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Services\Billing;

        final class BillingService {}
        PHP, $updated);
    assertSameValue('App\Services\Billing\BillingService', $decoded['symbolMappings'][0]['newSymbol']);
});

test('analyze command updates imported short references when class basenames change', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Shared/RichText/Ui/Web', 0o777, true);
    \mkdir($root . '/tests/Shared/RichText', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
        'autoload-dev' => [
            'psr-4' => [
                'App\\Tests\\' => 'tests/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Shared/RichText/Ui/Web/RichTextBlockWebRenderer.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web;

        interface RichTextBlockWebRenderer {}
        PHP);
    $renderer = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web;

        use App\Shared\RichText\Ui\Web\RichTextBlockWebRenderer;

        final class HtmlRichTextBlockRenderer implements RichTextBlockWebRenderer
        {
            public function render(RichTextBlockWebRenderer $renderer): RichTextBlockWebRenderer
            {
                return $renderer;
            }
        }
        PHP;
    \file_put_contents($root . '/src/Shared/RichText/Ui/Web/HtmlRichTextBlockRenderer.php', $renderer);
    $test = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Shared\RichText;

        use App\Shared\RichText\Ui\Web\HtmlRichTextBlockRenderer;

        $renderer = new HtmlRichTextBlockRenderer();
        $matches = $renderer instanceof HtmlRichTextBlockRenderer;
        HtmlRichTextBlockRenderer::make();
        PHP;
    \file_put_contents($root . '/tests/Shared/RichText/RendererTest.php', $test);
    $services = <<<'PHP'
        <?php

        use App\Shared\RichText\Ui\Web\HtmlRichTextBlockRenderer;

        return static function ($services): void {
            $services->set(HtmlRichTextBlockRenderer::class);
        };
        PHP;
    \file_put_contents($root . '/services.php', $services);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Shared/RichText/Ui/Web',
        'newPath' => 'src/Shared/RichText/Ui/Web',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Shared/RichText/Ui/Web/RichTextBlockWebRenderer.php',
            'newPath' => 'src/Shared/RichText/Ui/Web/RichTextRenderableWebRenderer.php',
            'tracked' => true,
        ], [
            'oldPath' => 'src/Shared/RichText/Ui/Web/HtmlRichTextBlockRenderer.php',
            'newPath' => 'src/Shared/RichText/Ui/Web/HtmlRichTextRenderableRenderer.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updatedRenderer = apply_replacements_for_file(
        $renderer,
        $decoded['replacements'],
        'src/Shared/RichText/Ui/Web/HtmlRichTextBlockRenderer.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web;

        final class HtmlRichTextRenderableRenderer implements RichTextRenderableWebRenderer
        {
            public function render(RichTextRenderableWebRenderer $renderer): RichTextRenderableWebRenderer
            {
                return $renderer;
            }
        }
        PHP, $updatedRenderer);

    $updatedTest = apply_replacements_for_file($test, $decoded['replacements'], 'tests/Shared/RichText/RendererTest.php');
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Shared\RichText;

        use App\Shared\RichText\Ui\Web\HtmlRichTextRenderableRenderer;

        $renderer = new HtmlRichTextRenderableRenderer();
        $matches = $renderer instanceof HtmlRichTextRenderableRenderer;
        HtmlRichTextRenderableRenderer::make();
        PHP, $updatedTest);

    $updatedServices = apply_replacements_for_file($services, $decoded['replacements'], 'services.php');
    assertSameValue(<<<'PHP'
        <?php

        use App\Shared\RichText\Ui\Web\HtmlRichTextRenderableRenderer;

        return static function ($services): void {
            $services->set(HtmlRichTextRenderableRenderer::class);
        };
        PHP, $updatedServices);
});

test('analyze command updates imported short references when namespace and basename change', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Shared/RichText/Ui/Web/Block', 0o777, true);
    \mkdir($root . '/tests/Shared/RichText', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
        'autoload-dev' => [
            'psr-4' => [
                'App\\Tests\\' => 'tests/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));

    $blockRenderer = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web\Block;

        final class AccordionBlockWebRenderer
        {
            public static function make(): self
            {
                return new self();
            }
        }
        PHP;
    \file_put_contents($root . '/src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php', $blockRenderer);

    $consumer = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web;

        use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

        final class HtmlRichTextRenderer
        {
            private ?AccordionBlockWebRenderer $renderer = null;

            public function render(AccordionBlockWebRenderer $renderer): AccordionBlockWebRenderer
            {
                $this->renderer = $renderer;

                if (!$renderer instanceof AccordionBlockWebRenderer) {
                    return new AccordionBlockWebRenderer();
                }

                return AccordionBlockWebRenderer::make();
            }
        }
        PHP;
    \file_put_contents($root . '/src/Shared/RichText/Ui/Web/HtmlRichTextRenderer.php', $consumer);

    $services = <<<'PHP'
        <?php

        use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

        return static function ($services): void {
            $services->instanceof(AccordionBlockWebRenderer::class);
            $services->set(AccordionBlockWebRenderer::class);
        };
        PHP;
    \file_put_contents($root . '/services.php', $services);

    $test = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Shared\RichText;

        use App\Shared\RichText\Ui\Web\Block\AccordionBlockWebRenderer;

        $renderer = new AccordionBlockWebRenderer();
        $matches = $renderer instanceof AccordionBlockWebRenderer;
        PHP;
    \file_put_contents($root . '/tests/Shared/RichText/AccordionRendererTest.php', $test);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php',
        'newPath' => 'src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php',
            'newPath' => 'src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updatedRenderer = apply_replacements_for_file(
        $blockRenderer,
        $decoded['replacements'],
        'src/Shared/RichText/Ui/Web/Block/AccordionBlockWebRenderer.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web\Renderer;

        final class AccordionRenderableWebRenderer
        {
            public static function make(): self
            {
                return new self();
            }
        }
        PHP, $updatedRenderer);

    $updatedConsumer = apply_replacements_for_file(
        $consumer,
        $decoded['replacements'],
        'src/Shared/RichText/Ui/Web/HtmlRichTextRenderer.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web;

        use App\Shared\RichText\Ui\Web\Renderer\AccordionRenderableWebRenderer;

        final class HtmlRichTextRenderer
        {
            private ?AccordionRenderableWebRenderer $renderer = null;

            public function render(AccordionRenderableWebRenderer $renderer): AccordionRenderableWebRenderer
            {
                $this->renderer = $renderer;

                if (!$renderer instanceof AccordionRenderableWebRenderer) {
                    return new AccordionRenderableWebRenderer();
                }

                return AccordionRenderableWebRenderer::make();
            }
        }
        PHP, $updatedConsumer);

    $updatedServices = apply_replacements_for_file($services, $decoded['replacements'], 'services.php');
    assertSameValue(<<<'PHP'
        <?php

        use App\Shared\RichText\Ui\Web\Renderer\AccordionRenderableWebRenderer;

        return static function ($services): void {
            $services->instanceof(AccordionRenderableWebRenderer::class);
            $services->set(AccordionRenderableWebRenderer::class);
        };
        PHP, $updatedServices);

    $updatedTest = apply_replacements_for_file($test, $decoded['replacements'], 'tests/Shared/RichText/AccordionRendererTest.php');
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Shared\RichText;

        use App\Shared\RichText\Ui\Web\Renderer\AccordionRenderableWebRenderer;

        $renderer = new AccordionRenderableWebRenderer();
        $matches = $renderer instanceof AccordionRenderableWebRenderer;
        PHP, $updatedTest);
});

test('analyze command updates imported enum case references when basename changes', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Shared/RichText/Application', 0o777, true);
    \mkdir($root . '/platform/src/Shared/RichText/Ui/Web/Renderer', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));

    $enum = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Application;

        enum RichTextComponentKind
        {
            case Accordion;
            case Card;
        }
        PHP;
    \file_put_contents($root . '/platform/src/Shared/RichText/Application/RichTextComponentKind.php', $enum);

    $renderer = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web\Renderer;

        use App\Shared\RichText\Application\RichTextComponentKind;

        final class AccordionRenderableWebRenderer
        {
            public function kind(): RichTextComponentKind
            {
                return RichTextComponentKind::Accordion;
            }
        }
        PHP;
    \file_put_contents($root . '/platform/src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php', $renderer);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Shared/RichText/Application/RichTextComponentKind.php',
        'newPath' => 'platform/src/Shared/RichText/Application/RichTextDirectiveKind.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Shared/RichText/Application/RichTextComponentKind.php',
            'newPath' => 'platform/src/Shared/RichText/Application/RichTextDirectiveKind.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updatedEnum = apply_replacements_for_file(
        $enum,
        $decoded['replacements'],
        'platform/src/Shared/RichText/Application/RichTextComponentKind.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Application;

        enum RichTextDirectiveKind
        {
            case Accordion;
            case Card;
        }
        PHP, $updatedEnum);

    $updatedRenderer = apply_replacements_for_file(
        $renderer,
        $decoded['replacements'],
        'platform/src/Shared/RichText/Ui/Web/Renderer/AccordionRenderableWebRenderer.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Web\Renderer;

        use App\Shared\RichText\Application\RichTextDirectiveKind;

        final class AccordionRenderableWebRenderer
        {
            public function kind(): RichTextDirectiveKind
            {
                return RichTextDirectiveKind::Accordion;
            }
        }
        PHP, $updatedRenderer);
});

test('analyze command preserves old namespace dependencies in moved files', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceFilter.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceFilter {}
        PHP);
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceTotals.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceTotals {}
        PHP);
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceBatch
        {
            public function project(InvoiceFilter $range): InvoiceTotals
            {
                return new InvoiceTotals();
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceBatch.php',
            'php-namespace-declaration',
            'App\\Billing\\Archive\\Domain',
        ),
        'expected moved namespace replacement',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceBatch.php',
            'php-namespace-local-import',
            "use App\\Billing\\Domain\\InvoiceFilter;\nuse App\\Billing\\Domain\\InvoiceTotals;\n\n",
        ),
        'expected imports for short old-namespace dependencies',
    );
});

test('analyze command keeps imported short style and removes same namespace imports after class move', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);
    \mkdir($root . '/platform/src/Billing/Archive/Domain', 0o777, true);
    \mkdir($root . '/platform/src/Billing/Archive/Detailed/Application', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Archive/Domain/InvoiceLineCollection.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Archive\Domain;

        final class InvoiceLineCollection {}
        PHP);
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        use App\Billing\Archive\Domain\InvoiceLineCollection;

        final class InvoiceBatch
        {
            public function __construct(
                private ?InvoiceLineCollection $documents = null,
            ) {}
        }
        PHP);
    \file_put_contents($root . '/platform/src/Billing/Archive/Detailed/Application/ResolveDocument.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Archive\Detailed\Application;

        use App\Billing\Domain\InvoiceBatch;

        final class ResolveDocument
        {
            public function matches(object $changes): bool
            {
                if (!$changes instanceof InvoiceBatch) {
                    return false;
                }

                return new InvoiceBatch() instanceof InvoiceBatch;
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Archive/Detailed/Application/ResolveDocument.php',
            'php-use-statement',
            'use App\\Billing\\Archive\\Domain\\InvoiceBatch;',
        ),
        'expected updated import for moved symbol',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Archive/Detailed/Application/ResolveDocument.php',
            'php-class-name-reference',
            'InvoiceBatch',
        ),
        'expected instanceof/new expressions to stay short',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceBatch.php',
            'php-use-statement',
            '',
        ),
        'expected same-namespace import removal in moved file',
    );
});

test('analyze command adds imports for same namespace consumers of moved symbols', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceBatch {}
        PHP);
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceArchive.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceArchive
        {
            public function hasChanges(?InvoiceBatch $changes): bool
            {
                return $changes instanceof InvoiceBatch;
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceArchive.php',
            'php-namespace-local-import',
            "use App\\Billing\\Archive\\Domain\\InvoiceBatch;\n\n",
        ),
        'expected import insertion for same namespace consumer',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceArchive.php',
            'php-method-parameter-type',
            'InvoiceBatch',
        ),
        'expected nullable parameter type to stay short',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/Domain/InvoiceArchive.php',
            'php-class-name-reference',
            'InvoiceBatch',
        ),
        'expected instanceof expression to stay short',
    );
});

test('analyze command applies moved file imports before class declarations', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);
    \mkdir($root . '/platform/src/Billing/Archive/Domain', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    $original = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        use App\Billing\Archive\Domain\InvoiceLineCollection;

        final readonly class InvoiceBatch
        {
            public function __construct(
                public string $edition,
                public InvoiceFilter $range,
                public InvoiceTotals $stats,
                public InvoiceLineCollection $documents,
            ) {}
        }
        PHP;
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', $original);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updated = apply_replacements_for_file($original, $decoded['replacements'], 'platform/src/Billing/Domain/InvoiceBatch.php');
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Archive\Domain;

        use App\Billing\Domain\InvoiceFilter;
        use App\Billing\Domain\InvoiceTotals;

        final readonly class InvoiceBatch
        {
            public function __construct(
                public string $edition,
                public InvoiceFilter $range,
                public InvoiceTotals $stats,
                public InvoiceLineCollection $documents,
            ) {}
        }
        PHP, $updated);
});

test('analyze command keeps same file helper classes namespace local after a move', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/tests/Application/Billing/Document', 0o777, true);
    \mkdir($root . '/platform/tests/Billing/Archive/Detailed/Application', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload-dev' => [
            'psr-4' => [
                'App\\Tests\\' => 'tests/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    $original = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Application\Billing\Invoice;

        final class RewriteInvoiceRichTextLinksTest
        {
            public function helper(): Helper
            {
                return new Helper();
            }
        }

        final class Helper {}
        PHP;
    \file_put_contents($root . '/platform/tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php', $original);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php',
        'newPath' => 'platform/tests/Billing/Archive/Detailed/Application/RewriteInvoiceRichTextLinksTest.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php',
            'newPath' => 'platform/tests/Billing/Archive/Detailed/Application/RewriteInvoiceRichTextLinksTest.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updated = apply_replacements_for_file(
        $original,
        $decoded['replacements'],
        'platform/tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php',
    );
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Billing\Archive\Detailed\Application;

        final class RewriteInvoiceRichTextLinksTest
        {
            public function helper(): Helper
            {
                return new Helper();
            }
        }

        final class Helper {}
        PHP, $updated);
});

test('analyze command updates twig component yaml namespace and template directory references', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/FileTree/Ui/Web/Twig', 0o777, true);
    \mkdir($root . '/platform/src/Billing/Reminder/Ui/Web/Twig', 0o777, true);
    \mkdir($root . '/platform/config/packages', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/config/packages/twig.yaml', <<<'YAML'
        twig:
          paths:
            '%kernel.project_dir%/src/Billing': Billing
        YAML);
    $originalConfig = <<<'YAML'
        twig_component:
          defaults:
            'App\Billing\FileTree\Ui\Web\':
              template_directory: '@Billing/FileTree/Ui/Web/Twig'
        YAML;
    \file_put_contents($root . '/platform/config/packages/twig_component.yaml', $originalConfig);
    $originalAssetConfig = <<<'YAML'
        framework:
          asset_mapper:
            paths:
              - 'src/Billing/FileTree/Ui/Web/'
        YAML;
    \file_put_contents($root . '/platform/config/packages/asset_mapper.yaml', $originalAssetConfig);
    \file_put_contents($root . '/platform/src/Billing/FileTree/Ui/Web/FileTreeComponent.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\FileTree\Ui\Web;

        #[AsTwigComponent(template: '@Billing/FileTree/Ui/Web/Twig/file-tree.html.twig')]
        final class FileTreeComponent {}
        PHP);
    \file_put_contents($root . '/platform/src/Billing/FileTree/Ui/Web/Twig/file-tree.html.twig', '<div></div>');

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/FileTree/Ui/Web',
        'newPath' => 'platform/src/Billing/Reminder/Ui/Web',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/FileTree/Ui/Web/FileTreeComponent.php',
            'newPath' => 'platform/src/Billing/Reminder/Ui/Web/FileTreeComponent.php',
            'tracked' => true,
        ], [
            'oldPath' => 'platform/src/Billing/FileTree/Ui/Web/Twig/file-tree.html.twig',
            'newPath' => 'platform/src/Billing/Reminder/Ui/Web/Twig/file-tree.html.twig',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => true,
        ],
    ]);

    $updatedConfig = apply_replacements_for_file(
        $originalConfig,
        $decoded['replacements'],
        'platform/config/packages/twig_component.yaml',
    );
    assertSameValue(<<<'YAML'
        twig_component:
          defaults:
            'App\Billing\Reminder\Ui\Web\':
              template_directory: '@Billing/Reminder/Ui/Web/Twig'
        YAML, $updatedConfig);
    $updatedAssetConfig = apply_replacements_for_file(
        $originalAssetConfig,
        $decoded['replacements'],
        'platform/config/packages/asset_mapper.yaml',
    );
    assertSameValue(<<<'YAML'
        framework:
          asset_mapper:
            paths:
              - 'src/Billing/Reminder/Ui/Web/'
        YAML, $updatedAssetConfig);
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/src/Billing/FileTree/Ui/Web/FileTreeComponent.php',
            'twigcomponenttemplateattributereplacementrule',
            "'@Billing/Reminder/Ui/Web/Twig/file-tree.html.twig'",
        ),
        'expected Twig component template attribute rewrite',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'platform/config/packages/asset_mapper.yaml',
            'yaml-asset-mapper-path',
            "'src/Billing/Reminder/Ui/Web/'",
        ),
        'expected asset mapper path rewrite',
    );
    assertTrueValue(
        has_path_mapping(
            $decoded['pathMappings'],
            'project-path-directory',
            'src/Billing/FileTree/Ui/Web/',
            'src/Billing/Reminder/Ui/Web/',
        ),
        'expected reported path mapping for asset mapper rewrite',
    );
});

test('analyze command applies consumer imports inside the import block before interfaces', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Domain', 0o777, true);

    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatch.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceBatch {}
        PHP);
    $original = <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        use App\Customer\Domain\CustomerId;

        interface InvoiceBatchRepository
        {
            public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
        }
        PHP;
    \file_put_contents($root . '/platform/src/Billing/Domain/InvoiceBatchRepository.php', $original);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
        'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Domain/InvoiceBatch.php',
            'newPath' => 'platform/src/Billing/Archive/Domain/InvoiceBatch.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    $updated = apply_replacements_for_file($original, $decoded['replacements'], 'platform/src/Billing/Domain/InvoiceBatchRepository.php');
    assertSameValue(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        use App\Customer\Domain\CustomerId;
        use App\Billing\Archive\Domain\InvoiceBatch;

        interface InvoiceBatchRepository
        {
            public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
        }
        PHP, $updated);
});

test('analyze command uses imports when fully qualified type usage duplicates an import', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Billing/Domain/Archive', 0o777, true);
    \mkdir($root . '/src/Consumer', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Billing/Domain/Archive/InvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain\Archive;

        final class InvoiceLine {}
        PHP);
    \file_put_contents($root . '/src/Consumer/UsesInvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Consumer;

        use App\Billing\Domain\Archive\InvoiceLine;

        final class UsesInvoiceLine
        {
            public function make(): \App\Billing\Domain\Archive\InvoiceLine
            {
                return new \App\Billing\Domain\Archive\InvoiceLine();
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Billing/Domain/Archive',
        'newPath' => 'src/Billing/Archive/Domain',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Billing/Domain/Archive/InvoiceLine.php',
            'newPath' => 'src/Billing/Archive/Domain/InvoiceLine.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Consumer/UsesInvoiceLine.php',
            'php-method-return-type',
            'InvoiceLine',
        ),
        'expected explicit fully qualified return type to use the import',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Consumer/UsesInvoiceLine.php',
            'php-fully-qualified-class-name',
            'InvoiceLine',
        ),
        'expected explicit fully qualified constructor call to use the import',
    );
});

test('analyze command preserves aliased import type style', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Billing/Domain/Archive', 0o777, true);
    \mkdir($root . '/src/Consumer', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Billing/Domain/Archive/InvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain\Archive;

        final class InvoiceLine {}
        PHP);
    \file_put_contents($root . '/src/Consumer/UsesInvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Consumer;

        use App\Billing\Domain\Archive\InvoiceLine as SnapshotDocument;

        final class UsesInvoiceLine
        {
            public function make(): SnapshotDocument
            {
                return new SnapshotDocument();
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Billing/Domain/Archive',
        'newPath' => 'src/Billing/Archive/Domain',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Billing/Domain/Archive/InvoiceLine.php',
            'newPath' => 'src/Billing/Archive/Domain/InvoiceLine.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Consumer/UsesInvoiceLine.php',
            'php-use-statement',
            'use App\\Billing\\Archive\\Domain\\InvoiceLine as SnapshotDocument;',
        ),
        'expected aliased import target to update',
    );
    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Consumer/UsesInvoiceLine.php',
            'php-method-return-type',
            'SnapshotDocument',
        ),
        'expected aliased return type to stay aliased',
    );
});

test('analyze command removes redundant import when moved files land in same namespace', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Billing/Domain/Archive', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Billing/Domain/Archive/InvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain\Archive;

        final class InvoiceLine {}
        PHP);
    \file_put_contents($root . '/src/Billing/Domain/Archive/UsesInvoiceLine.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain\Archive;

        use App\Billing\Domain\Archive\InvoiceLine;

        final class UsesInvoiceLine
        {
            public function make(): InvoiceLine
            {
                return new InvoiceLine();
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Billing/Domain/Archive',
        'newPath' => 'src/Billing/Archive/Domain',
        'dryRun' => true,
        'moves' => [
            [
                'oldPath' => 'src/Billing/Domain/Archive/InvoiceLine.php',
                'newPath' => 'src/Billing/Archive/Domain/InvoiceLine.php',
                'tracked' => true,
            ],
            [
                'oldPath' => 'src/Billing/Domain/Archive/UsesInvoiceLine.php',
                'newPath' => 'src/Billing/Archive/Domain/UsesInvoiceLine.php',
                'tracked' => true,
            ],
        ],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Billing/Domain/Archive/UsesInvoiceLine.php',
            'php-use-statement',
            '',
        ),
        'expected redundant import removal after directory move',
    );
});

test('analyze command warns about string literals containing moved php symbols', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Billing/Archive/Infrastructure', 0o777, true);
    \mkdir($root . '/src/Billing/Archive/Core/Infrastructure', 0o777, true);
    \mkdir($root . '/tests/Architecture', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
        'autoload-dev' => [
            'psr-4' => [
                'App\\Tests\\' => 'tests/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Billing/Archive/Infrastructure/ArchiveProjector.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Archive\Infrastructure;

        final class ArchiveProjector {}
        PHP);
    \file_put_contents($root . '/tests/Architecture/ArchitectureDependencyRuleTest.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Tests\Architecture;

        final class ArchitectureDependencyRuleTest
        {
            public function testDiagnostic(): void
            {
                $expected = 'App\Billing\Archive\Infrastructure\ArchiveProjector must not be used here';
            }
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Billing/Archive/Infrastructure/ArchiveProjector.php',
        'newPath' => 'src/Billing/Archive/Core/Infrastructure/ArchiveProjector.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Billing/Archive/Infrastructure/ArchiveProjector.php',
            'newPath' => 'src/Billing/Archive/Core/Infrastructure/ArchiveProjector.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_warning(
            $decoded['warnings'],
            'tests/Architecture/ArchitectureDependencyRuleTest.php',
            'String literal references a moved PHP symbol; not changed.',
        ),
        'expected report-only warning for moved FQCN inside string literal',
    );
    assertSameValue(
        0,
        \count(\array_filter(
            $decoded['replacements'],
            static fn(array $replacement): bool => 'tests/Architecture/ArchitectureDependencyRuleTest.php' === $replacement['file'],
        )),
        'expected no string literal replacement by default',
    );
});

test('analyze command reports semantic rename hints without applying them', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/src/Shared/RichText/Ui/Rendering/CommonMark/Component', 0o777, true);
    \mkdir($root . '/src/Shared/RichText/Ui/Rendering/CommonMark/Directive', 0o777, true);
    \mkdir($root . '/config/packages', 0o777, true);

    \file_put_contents($root . '/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/src/Shared/RichText/Ui/Rendering/CommonMark/Component/ComponentRenderer.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Rendering\CommonMark\Component;

        interface ComponentRenderer {}
        PHP);
    \file_put_contents($root . '/src/Shared/RichText/Ui/Rendering/CommonMark/Directive/DirectiveNodeRenderer.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Shared\RichText\Ui\Rendering\CommonMark\Directive;

        final class DirectiveNodeRenderer
        {
            public function __construct(private iterable $componentRenderers) {}

            public function tag(): string
            {
                return 'app.rich_text_component_renderer';
            }
        }
        PHP);
    \file_put_contents($root . '/config/packages/services.yaml', <<<'YAML'
        services:
          App\Shared\RichText\:
            tags: ['app.rich_text_component_renderer']
        YAML);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'src/Shared/RichText/Ui/Rendering/CommonMark/Component/ComponentRenderer.php',
        'newPath' => 'src/Shared/RichText/Ui/Rendering/CommonMark/Directive/DirectiveRenderer.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'src/Shared/RichText/Ui/Rendering/CommonMark/Component/ComponentRenderer.php',
            'newPath' => 'src/Shared/RichText/Ui/Rendering/CommonMark/Directive/DirectiveRenderer.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertTrueValue(
        has_warning(
            $decoded['warnings'],
            'src/Shared/RichText/Ui/Rendering/CommonMark/Directive/DirectiveNodeRenderer.php',
            'Semantic name "componentRenderers" resembles moved symbol; consider "directiveRenderers". Not changed.',
        ),
        'expected report-only variable hint',
    );
    assertTrueValue(
        has_warning(
            $decoded['warnings'],
            'src/Shared/RichText/Ui/Rendering/CommonMark/Directive/DirectiveNodeRenderer.php',
            'Semantic name "component_renderer" resembles moved symbol; consider "app.rich_text_directive_renderer". Not changed.',
        ),
        'expected report-only PHP string hint',
    );
    assertTrueValue(
        has_warning(
            $decoded['warnings'],
            'config/packages/services.yaml',
            'Semantic name "component_renderer" resembles moved symbol; consider "directive_renderer". Not changed.',
        ),
        'expected report-only config hint',
    );
    assertSameValue(
        0,
        \count(\array_filter(
            $decoded['replacements'],
            static fn(array $replacement): bool => 'config/packages/services.yaml' === $replacement['file'],
        )),
        'expected no semantic config replacement by default',
    );
});

test('analyze command skips configured fixture paths during semantic rewrites', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-analyze-' . \uniqid();
    \mkdir($root . '/platform/src/Billing/Archive/Infrastructure', 0o777, true);
    \mkdir($root . '/platform/src/Billing/Archive/Core/Infrastructure', 0o777, true);
    \mkdir($root . '/platform/local/phpstan/tests/fixtures', 0o777, true);

    \file_put_contents($root . '/platform/.refactorlah.json', \json_encode([
        'exclude' => [
            'local/phpstan/tests/fixtures/**',
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/composer.json', \json_encode([
        'autoload' => [
            'psr-4' => [
                'App\\' => 'src/',
            ],
        ],
    ], JSON_PRETTY_PRINT | JSON_THROW_ON_ERROR));
    \file_put_contents($root . '/platform/src/Billing/Archive/Infrastructure/ArchiveProjector.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Archive\Infrastructure;

        final class ArchiveProjector {}
        PHP);
    \file_put_contents($root . '/platform/local/phpstan/tests/fixtures/ArchitectureDependency.php', <<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Local\Phpstan\Tests\Fixtures;

        use App\Billing\Archive\Infrastructure\ArchiveProjector;

        final class ArchitectureDependency
        {
            public function project(ArchiveProjector $projector): void {}
        }
        PHP);

    $decoded = run_adapter($root, [
        'protocolVersion' => 1,
        'projectRoot' => '.',
        'oldPath' => 'platform/src/Billing/Archive/Infrastructure/ArchiveProjector.php',
        'newPath' => 'platform/src/Billing/Archive/Core/Infrastructure/ArchiveProjector.php',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'platform/src/Billing/Archive/Infrastructure/ArchiveProjector.php',
            'newPath' => 'platform/src/Billing/Archive/Core/Infrastructure/ArchiveProjector.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => false,
        ],
    ]);

    assertSameValue(
        0,
        \count(\array_filter(
            $decoded['replacements'],
            static fn(array $replacement): bool => 'platform/local/phpstan/tests/fixtures/ArchitectureDependency.php' === $replacement['file'],
        )),
        'expected no fixture replacements for configured exclude',
    );
});

test('analyze command rejects invalid protocol metadata', function (): void
{
    $repoRoot = \dirname(__DIR__, 4);
    $fixtureRoot = $repoRoot . '/adapters/php/tests/fixtures/php-basic';

    $decoded = run_adapter($fixtureRoot, [
        'protocolVersion' => 2,
        'projectRoot' => '.',
        'oldPath' => 'app/Services/Billing',
        'newPath' => 'app/Domain/Billing',
        'dryRun' => true,
        'moves' => [[
            'oldPath' => 'app/Services/Billing/InvoiceService.php',
            'newPath' => 'app/Domain/Billing/InvoiceService.php',
            'tracked' => true,
        ]],
        'options' => [
            'includePhp' => true,
            'includeTwig' => true,
        ],
    ]);

    assertSameValue('adapter request must use protocolVersion 1', $decoded['errors'][0] ?? null);
});

/**
 * @param array<string,mixed> $request
 * @return array{
 *   protocolVersion:int,
 *   adapter:string,
 *   symbolMappings:list<array{
 *     kind:string,
 *     oldPath:string,
 *     newPath:string,
 *     oldSymbol:string,
 *     newSymbol:string,
 *     oldNamespace:string,
 *     newNamespace:string,
 *     shortName:string
 *   }>,
 *   pathMappings:list<array<string,mixed>>,
 *   replacements:list<array{
 *     file:string,
 *     start:int,
 *     end:int,
 *     replacement:string,
 *     reason:string,
 *     rule:string
 *   }>,
 *   warnings:list<array{message:string,file?:string,line?:int}>,
 *   errors:list<string>
 * }
 */
function run_adapter(string $projectRoot, array $request): array
{
    $repoRoot = \dirname(__DIR__, 4);
    $adapterBinary = $repoRoot . '/adapters/php/bin/refactorlah-php';
    $encoded = \json_encode($request, JSON_THROW_ON_ERROR);
    $command = \sprintf(
        'cd %s && printf %s | %s analyze',
        \escapeshellarg($projectRoot),
        \escapeshellarg($encoded),
        \escapeshellarg($adapterBinary)
    );
    $output = \shell_exec($command);
    if (!\is_string($output) || '' === $output) {
        throw new RuntimeException('expected adapter output');
    }
    $decoded = \json_decode($output, true);
    if (!\is_array($decoded)) {
        throw new RuntimeException('expected decoded adapter response array');
    }

    return normalize_adapter_response(normalize_string_key_array($decoded));
}

/**
 * @param list<array{
 *   file:string,
 *   start:int,
 *   end:int,
 *   replacement:string,
 *   reason:string,
 *   rule:string
 * }> $replacements
 */
function has_replacement(array $replacements, string $file, string $reason, string $replacement): bool
{
    foreach ($replacements as $candidate) {
        if ($candidate['file'] !== $file) {
            continue;
        }
        if ($candidate['reason'] !== $reason) {
            continue;
        }
        if ($candidate['replacement'] !== $replacement) {
            continue;
        }

        return true;
    }

    return false;
}

/**
 * @param list<array{message:string,file?:string,line?:int}> $warnings
 */
function has_warning(array $warnings, string $file, string $message): bool
{
    foreach ($warnings as $warning) {
        if (($warning['file'] ?? '') === $file && $warning['message'] === $message) {
            return true;
        }
    }

    return false;
}

/**
 * @param list<array<string,mixed>> $pathMappings
 */
function has_path_mapping(array $pathMappings, string $kind, string $oldReference, string $newReference): bool
{
    foreach ($pathMappings as $pathMapping) {
        if (($pathMapping['kind'] ?? null) === $kind
            && ($pathMapping['oldReference'] ?? null) === $oldReference
            && ($pathMapping['newReference'] ?? null) === $newReference) {
            return true;
        }
    }

    return false;
}

/**
 * @param list<array{
 *   file:string,
 *   start:int,
 *   end:int,
 *   replacement:string,
 *   reason:string,
 *   rule:string
 * }> $replacements
 */
function apply_replacements_for_file(string $content, array $replacements, string $file): string
{
    $filtered = [];
    foreach ($replacements as $replacement) {
        if ($replacement['file'] !== $file) {
            continue;
        }
        $filtered[] = $replacement;
    }

    \usort($filtered, static function (array $left, array $right): int
    {
        return $right['start'] <=> $left['start'];
    });

    foreach ($filtered as $replacement) {
        $content = \mb_substr($content, 0, $replacement['start'])
            . $replacement['replacement']
            . \mb_substr($content, $replacement['end']);
    }

    return $content;
}

/**
 * @param array<string,mixed> $decoded
 * @return array{
 *   protocolVersion:int,
 *   adapter:string,
 *   symbolMappings:list<array{
 *     kind:string,
 *     oldPath:string,
 *     newPath:string,
 *     oldSymbol:string,
 *     newSymbol:string,
 *     oldNamespace:string,
 *     newNamespace:string,
 *     shortName:string
 *   }>,
 *   pathMappings:list<array<string,mixed>>,
 *   replacements:list<array{
 *     file:string,
 *     start:int,
 *     end:int,
 *     replacement:string,
 *     reason:string,
 *     rule:string
 *   }>,
 *   warnings:list<array{message:string,file?:string,line?:int}>,
 *   errors:list<string>
 * }
 */
function normalize_adapter_response(array $decoded): array
{
    return [
        'protocolVersion' => mixed_int($decoded['protocolVersion'] ?? null),
        'adapter' => mixed_string($decoded['adapter'] ?? null),
        'symbolMappings' => normalize_symbol_mappings($decoded['symbolMappings'] ?? null),
        'pathMappings' => normalize_path_mappings($decoded['pathMappings'] ?? null),
        'replacements' => normalize_replacements($decoded['replacements'] ?? null),
        'warnings' => normalize_warnings($decoded['warnings'] ?? null),
        'errors' => normalize_string_list($decoded['errors'] ?? null),
    ];
}

/**
 * @param mixed $symbolMappings
 * @return list<array{
 *   kind:string,
 *   oldPath:string,
 *   newPath:string,
 *   oldSymbol:string,
 *   newSymbol:string,
 *   oldNamespace:string,
 *   newNamespace:string,
 *   shortName:string
 * }>
 */
function normalize_symbol_mappings(mixed $symbolMappings): array
{
    if (!\is_array($symbolMappings)) {
        return [];
    }

    $normalized = [];
    foreach ($symbolMappings as $mapping) {
        if (!\is_array($mapping)) {
            continue;
        }

        $normalized[] = [
            'kind' => mixed_string($mapping['kind'] ?? null),
            'oldPath' => mixed_string($mapping['oldPath'] ?? null),
            'newPath' => mixed_string($mapping['newPath'] ?? null),
            'oldSymbol' => mixed_string($mapping['oldSymbol'] ?? null),
            'newSymbol' => mixed_string($mapping['newSymbol'] ?? null),
            'oldNamespace' => mixed_string($mapping['oldNamespace'] ?? null),
            'newNamespace' => mixed_string($mapping['newNamespace'] ?? null),
            'shortName' => mixed_string($mapping['shortName'] ?? null),
        ];
    }

    return $normalized;
}

/**
 * @param mixed $pathMappings
 * @return list<array<string,mixed>>
 */
function normalize_path_mappings(mixed $pathMappings): array
{
    if (!\is_array($pathMappings)) {
        return [];
    }

    $normalized = [];
    foreach ($pathMappings as $mapping) {
        if (!\is_array($mapping)) {
            continue;
        }

        $entry = [];
        foreach ($mapping as $key => $value) {
            if (!\is_string($key)) {
                continue;
            }

            $entry[$key] = $value;
        }
        $normalized[] = $entry;
    }

    return $normalized;
}

/**
 * @param mixed $replacements
 * @return list<array{
 *   file:string,
 *   start:int,
 *   end:int,
 *   replacement:string,
 *   reason:string,
 *   rule:string
 * }>
 */
function normalize_replacements(mixed $replacements): array
{
    if (!\is_array($replacements)) {
        return [];
    }

    $normalized = [];
    foreach ($replacements as $replacement) {
        if (!\is_array($replacement)) {
            continue;
        }

        $normalized[] = [
            'file' => mixed_string($replacement['file'] ?? null),
            'start' => mixed_int($replacement['start'] ?? null),
            'end' => mixed_int($replacement['end'] ?? null),
            'replacement' => mixed_string($replacement['replacement'] ?? null),
            'reason' => mixed_string($replacement['reason'] ?? null),
            'rule' => mixed_string($replacement['rule'] ?? null),
        ];
    }

    return $normalized;
}

/**
 * @param mixed $warnings
 * @return list<array{message:string,file?:string,line?:int}>
 */
function normalize_warnings(mixed $warnings): array
{
    if (!\is_array($warnings)) {
        return [];
    }

    $normalized = [];
    foreach ($warnings as $warning) {
        if (!\is_array($warning)) {
            continue;
        }

        $entry = ['message' => mixed_string($warning['message'] ?? null)];
        if (\array_key_exists('file', $warning)) {
            $entry['file'] = mixed_string($warning['file']);
        }
        if (\array_key_exists('line', $warning)) {
            $entry['line'] = mixed_int($warning['line']);
        }

        $normalized[] = $entry;
    }

    return $normalized;
}

/**
 * @param mixed $strings
 * @return list<string>
 */
function normalize_string_list(mixed $strings): array
{
    if (!\is_array($strings)) {
        return [];
    }

    $normalized = [];
    foreach ($strings as $string) {
        $normalized[] = mixed_string($string);
    }

    return $normalized;
}

/**
 * @param array<mixed,mixed> $values
 * @return array<string,mixed>
 */
function normalize_string_key_array(array $values): array
{
    $normalized = [];
    foreach ($values as $key => $value) {
        if (!\is_string($key)) {
            continue;
        }

        $normalized[$key] = $value;
    }

    return $normalized;
}

function mixed_string(mixed $value): string
{
    return \is_string($value) ? $value : '';
}

function mixed_int(mixed $value): int
{
    return \is_int($value) ? $value : 0;
}
