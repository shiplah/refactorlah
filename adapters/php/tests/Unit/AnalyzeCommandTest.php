<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\AnalyzeCommand;

test('analyze command emits valid protocol response for fixture project', function (): void {
    $repoRoot = dirname(__DIR__, 4);
    $fixtureRoot = $repoRoot . '/tests/fixtures/php-basic';
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

    $encoded = json_encode($request, JSON_THROW_ON_ERROR);
    $command = sprintf(
        'cd %s && printf %s | %s analyze',
        escapeshellarg($fixtureRoot),
        escapeshellarg($encoded),
        escapeshellarg($adapterBinary)
    );
    $output = shell_exec($command);
    assertTrueValue(is_string($output) && $output !== '', 'expected adapter output');

    $decoded = json_decode($output, true);
    assertSameValue(1, $decoded['protocolVersion']);
    assertSameValue('php', $decoded['adapter']);
    assertTrueValue(count($decoded['symbolMappings']) >= 1, 'expected symbol mappings');
});
