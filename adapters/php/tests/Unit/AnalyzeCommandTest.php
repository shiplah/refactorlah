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
            "\n\nuse App\\Billing\\Domain\\InvoiceFilter;\nuse App\\Billing\\Domain\\InvoiceTotals;",
        ),
        'expected imports for short old-namespace dependencies',
    );
});

test('analyze command preserves explicit fully qualified type usage when imports also exist', function (): void
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

    assertTrueValue(
        has_replacement(
            $decoded['replacements'],
            'src/Consumer/UsesInvoiceLine.php',
            'php-method-return-type',
            '\\App\\Billing\\Archive\\Domain\\InvoiceLine',
        ),
        'expected explicit fully qualified return type to stay fully qualified',
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

/**
 * @param array<string,mixed> $request
 * @return array<string,mixed>
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
    assertTrueValue(\is_string($output) && '' !== $output, 'expected adapter output');

    /** @var array<string,mixed> $decoded */
    $decoded = \json_decode($output, true);

    return $decoded;
}

/**
 * @param list<array<string,mixed>> $replacements
 */
function has_replacement(array $replacements, string $file, string $reason, string $replacement): bool
{
    foreach ($replacements as $candidate) {
        if (($candidate['file'] ?? null) !== $file) {
            continue;
        }
        if (($candidate['reason'] ?? null) !== $reason) {
            continue;
        }
        if (($candidate['replacement'] ?? null) !== $replacement) {
            continue;
        }

        return true;
    }

    return false;
}
