<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\Composer\Psr4Map;
use Refactorlah\PhpAdapter\Php\PhpSymbolScanner;
use Refactorlah\PhpAdapter\Php\Psr4NamespaceResolver;
use Refactorlah\PhpAdapter\Protocol\MoveCollection;

test('php symbol scanner derives mapping for deterministic PSR-4 move', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-symbol-' . \uniqid();
    \mkdir($root . '/app/Services/Billing', 0o777, true);
    \file_put_contents($root . '/app/Services/Billing/InvoiceService.php', <<<'PHP'
        <?php
        namespace App\Services\Billing;
        final class InvoiceService {}
        PHP);

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    $result = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), MoveCollection::fromMixed([[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]));

    assertSameValue(1, \count($result->symbolMappings));
    assertSameValue(0, \count($result->warnings));
    assertSameValue('App\Services\Billing\InvoiceService', $result->symbolMappings[0]->oldSymbol);
    assertSameValue('App\Domain\Billing\InvoiceService', $result->symbolMappings[0]->newSymbol);
});

test('php symbol scanner warns for non-PSR-4 path', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-symbol-' . \uniqid();
    \mkdir($root . '/misc', 0o777, true);
    \file_put_contents($root . '/misc/InvoiceService.php', "<?php\nfinal class InvoiceService {}\n");

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    $result = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), MoveCollection::fromMixed([[
        'oldPath' => 'misc/InvoiceService.php',
        'newPath' => 'misc/Other.php',
        'tracked' => true,
    ]]));

    assertSameValue(0, \count($result->symbolMappings));
    assertSameValue(1, \count($result->warnings));
});

test('php symbol scanner warns when multiple top-level symbols are ambiguous', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-symbol-' . \uniqid();
    \mkdir($root . '/app/Services/Billing', 0o777, true);
    \file_put_contents($root . '/app/Services/Billing/InvoiceService.php', <<<'PHP'
        <?php
        namespace App\Services\Billing;
        final class A {}
        final class B {}
        PHP);

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    $result = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), MoveCollection::fromMixed([[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]));

    assertSameValue(0, \count($result->symbolMappings));
    assertSameValue(1, \count($result->warnings));
});

test('php symbol scanner prefers filename-matching symbol when multiple top-level symbols exist', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-symbol-' . \uniqid();
    \mkdir($root . '/app/Services/Billing', 0o777, true);
    \file_put_contents($root . '/app/Services/Billing/InvoiceService.php', <<<'PHP'
        <?php
        namespace App\Services\Billing;
        final class Helper {}
        final class InvoiceService {}
        PHP);

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    $result = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), MoveCollection::fromMixed([[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]));

    assertSameValue(1, \count($result->symbolMappings));
    assertSameValue(0, \count($result->warnings));
    assertSameValue('App\Services\Billing\InvoiceService', $result->symbolMappings[0]->oldSymbol);
});

test('php symbol scanner warns when file cannot be parsed', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-symbol-' . \uniqid();
    \mkdir($root . '/app/Services/Billing', 0o777, true);
    \file_put_contents($root . '/app/Services/Billing/InvoiceService.php', "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {\n");

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    $result = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), MoveCollection::fromMixed([[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]));

    assertSameValue(0, \count($result->symbolMappings));
    assertSameValue(1, \count($result->warnings));
    assertSameValue('PHP file could not be parsed; symbol mapping skipped.', $result->warnings[0]->message);
});
