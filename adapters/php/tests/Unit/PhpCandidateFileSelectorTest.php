<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\Php\PhpCandidateFileSelector;
use Refactorlah\PhpAdapter\Php\SymbolMapping;

test('php candidate selector keeps moved files and files that mention moved symbols', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-candidates-' . \uniqid();
    \mkdir($root . '/src/Billing/Domain/Archive', 0o777, true);
    \mkdir($root . '/src/Consumer', 0o777, true);

    \file_put_contents($root . '/src/Billing/Domain/Archive/InvoiceLine.php', <<<'PHP'
        <?php

        namespace App\Billing\Domain\Archive;

        final class InvoiceLine {}
        PHP);
    \file_put_contents($root . '/src/Consumer/UsesInvoiceLine.php', <<<'PHP'
        <?php

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
    \file_put_contents($root . '/src/Consumer/UnrelatedFile.php', <<<'PHP'
        <?php

        namespace App\Consumer;

        final class UnrelatedFile {}
        PHP);

    $mapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'src/Billing/Domain/Archive/InvoiceLine.php',
        newPath: 'src/Billing/Archive/Domain/InvoiceLine.php',
        oldSymbol: 'App\\Billing\\Domain\\Archive\\InvoiceLine',
        newSymbol: 'App\\Billing\\Archive\\Domain\\InvoiceLine',
        oldNamespace: 'App\\Billing\\Domain\\Archive',
        newNamespace: 'App\\Billing\\Archive\\Domain',
        shortName: 'InvoiceLine',
    );

    $selected = (new PhpCandidateFileSelector())->select(
        projectRoot: $root,
        files: [
            'src/Consumer/UnrelatedFile.php',
            'src/Consumer/UsesInvoiceLine.php',
            'src/Billing/Domain/Archive/InvoiceLine.php',
        ],
        symbolMappings: [$mapping],
        movedPhpFiles: ['src/Billing/Domain/Archive/InvoiceLine.php'],
    );

    assertSameValue([
        'src/Consumer/UsesInvoiceLine.php',
        'src/Billing/Domain/Archive/InvoiceLine.php',
    ], $selected);
});

test('php candidate selector skips parsing when no symbol mappings exist', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-php-candidates-' . \uniqid();
    \mkdir($root . '/src', 0o777, true);
    \file_put_contents($root . '/src/Anything.php', "<?php\n");

    $selected = (new PhpCandidateFileSelector())->select(
        projectRoot: $root,
        files: ['src/Anything.php'],
        symbolMappings: [],
        movedPhpFiles: ['src/Anything.php'],
    );

    assertSameValue([], $selected);
});
