<?php

declare(strict_types=1);

use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use Refactorlah\PhpAdapter\Composer\Psr4Map;
use Refactorlah\PhpAdapter\Php\PhpSymbolScanner;
use Refactorlah\PhpAdapter\Php\Psr4NamespaceResolver;

test('php symbol scanner derives mapping for deterministic PSR-4 move', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-php-symbol-' . uniqid();
    mkdir($root . '/app/Services/Billing', 0777, true);
    file_put_contents($root . '/app/Services/Billing/InvoiceService.php', <<<'PHP'
<?php
namespace App\Services\Billing;
final class InvoiceService {}
PHP);

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    [$mappings, $warnings] = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), [[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]);

    assertSameValue(1, count($mappings));
    assertSameValue(0, count($warnings));
    assertSameValue('App\Services\Billing\InvoiceService', $mappings[0]->oldSymbol);
    assertSameValue('App\Domain\Billing\InvoiceService', $mappings[0]->newSymbol);
});

test('php symbol scanner warns for non-PSR-4 path', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-php-symbol-' . uniqid();
    mkdir($root . '/misc', 0777, true);
    file_put_contents($root . '/misc/InvoiceService.php', "<?php\nfinal class InvoiceService {}\n");

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    [$mappings, $warnings] = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), [[
        'oldPath' => 'misc/InvoiceService.php',
        'newPath' => 'misc/Other.php',
        'tracked' => true,
    ]]);

    assertSameValue(0, count($mappings));
    assertSameValue(1, count($warnings));
});

test('php symbol scanner warns when multiple top-level symbols are ambiguous', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-php-symbol-' . uniqid();
    mkdir($root . '/app/Services/Billing', 0777, true);
    file_put_contents($root . '/app/Services/Billing/InvoiceService.php', <<<'PHP'
<?php
namespace App\Services\Billing;
final class A {}
final class B {}
PHP);

    $scanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
    [$mappings, $warnings] = $scanner->scan($root, new Psr4Map(['App\\' => ['app']]), [[
        'oldPath' => 'app/Services/Billing/InvoiceService.php',
        'newPath' => 'app/Domain/Billing/InvoiceService.php',
        'tracked' => true,
    ]]);

    assertSameValue(0, count($mappings));
    assertSameValue(1, count($warnings));
});
