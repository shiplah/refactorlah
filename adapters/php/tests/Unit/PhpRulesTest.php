<?php

declare(strict_types=1);

use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\SymbolMapping;

function php_context(string $content, string $path = 'app/Http/Controllers/InvoiceController.php'): PhpFileContext
{
    $parser = (new ParserFactory())->createForNewestSupportedVersion();
    $ast = array_values($parser->parse($content) ?? []);
    $traverser = new NodeTraverser();
    $traverser->addVisitor(new NameResolver(options: ['preserveOriginalNames' => true]));
    $resolved = array_values($traverser->traverse($ast));
    \Refactorlah\PhpAdapter\Php\RuleSupport::attachParents($resolved);
    return new PhpFileContext($path, $content, $resolved);
}

function php_analysis_context(): AnalysisContext
{
    $mapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'app/Services/Billing/InvoiceService.php',
        newPath: 'app/Domain/Billing/InvoiceService.php',
        oldSymbol: 'App\Services\Billing\InvoiceService',
        newSymbol: 'App\Domain\Billing\InvoiceService',
        oldNamespace: 'App\Services\Billing',
        newNamespace: 'App\Domain\Billing',
        shortName: 'InvoiceService',
    );

    return new AnalysisContext([$mapping->oldSymbol => $mapping]);
}

function php_analysis_context_for_moved_consumer(): AnalysisContext
{
    $invoiceMapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'app/Services/Billing/InvoiceService.php',
        newPath: 'app/Domain/Billing/InvoiceService.php',
        oldSymbol: 'App\Services\Billing\InvoiceService',
        newSymbol: 'App\Domain\Billing\InvoiceService',
        oldNamespace: 'App\Services\Billing',
        newNamespace: 'App\Domain\Billing',
        shortName: 'InvoiceService',
    );
    $consumerMapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'app/Services/Billing/Consumer.php',
        newPath: 'app/Domain/Billing/Consumer.php',
        oldSymbol: 'App\Services\Billing\Consumer',
        newSymbol: 'App\Domain\Billing\Consumer',
        oldNamespace: 'App\Services\Billing',
        newNamespace: 'App\Domain\Billing',
        shortName: 'Consumer',
    );

    return new AnalysisContext([
        $invoiceMapping->oldSymbol => $invoiceMapping,
        $consumerMapping->oldSymbol => $consumerMapping,
    ]);
}

function php_analysis_context_for_namespace_local_dependency_move(): AnalysisContext
{
    $mapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'src/Billing/Domain/InvoiceBatch.php',
        newPath: 'src/Billing/Archive/Domain/InvoiceBatch.php',
        oldSymbol: 'App\Billing\Domain\InvoiceBatch',
        newSymbol: 'App\Billing\Archive\Domain\InvoiceBatch',
        oldNamespace: 'App\Billing\Domain',
        newNamespace: 'App\Billing\Archive\Domain',
        shortName: 'InvoiceBatch',
    );

    return new AnalysisContext([$mapping->oldSymbol => $mapping]);
}

test('namespace declaration rule updates moved file namespace', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\NamespaceDeclarationReplacementRule();
    $context = php_context("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n", 'app/Services/Billing/InvoiceService.php');
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('App\Domain\Billing', $replacements[0]->replacement);
});

test('use statement rule updates imported symbol', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\UseStatementReplacementRule();
    $context = php_context("<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('use App\\Domain\\Billing\\InvoiceService;', $replacements[0]->replacement);
});

test('use statement rule removes import when moved file now shares namespace', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\UseStatementReplacementRule();
    $context = php_context(
        "<?php\nnamespace App\\Services\\Billing;\n\nuse App\\Services\\Billing\\InvoiceService;\n\nfinal class Consumer {}\n",
        'app/Services/Billing/Consumer.php',
    );
    $replacements = $rule->collect($context, php_analysis_context_for_moved_consumer());
    assertSameValue(1, \count($replacements));
    assertSameValue('', $replacements[0]->replacement);
});

test('use statement rule removes same namespace import when only the current file moves', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\UseStatementReplacementRule();
    $context = php_context(
        "<?php\nnamespace App\\Billing\\Domain;\n\nuse App\\Billing\\Archive\\Domain\\InvoiceLineCollection;\n\nfinal class InvoiceBatch {}\n",
        'src/Billing/Domain/InvoiceBatch.php',
    );
    $replacements = $rule->collect($context, php_analysis_context_for_namespace_local_dependency_move());
    assertSameValue(1, \count($replacements));
    assertSameValue('', $replacements[0]->replacement);
});

test('namespace local dependency import rule preserves short type references in moved files', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\NamespaceLocalDependencyImportRule();
    $context = php_context(<<<'PHP'
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
        PHP, 'src/Billing/Domain/InvoiceBatch.php');

    $replacements = $rule->collect($context, php_analysis_context_for_namespace_local_dependency_move());
    assertSameValue(1, \count($replacements));
    assertSameValue(
        "use App\\Billing\\Domain\\InvoiceFilter;\nuse App\\Billing\\Domain\\InvoiceTotals;\n\n",
        $replacements[0]->replacement,
    );
});

test('namespace local dependency import rule keeps same file helper classes namespace local after a move', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\NamespaceLocalDependencyImportRule();
    $context = php_context(<<<'PHP'
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
        PHP, 'tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php');
    $mapping = new SymbolMapping(
        kind: 'class',
        oldPath: 'tests/Application/Billing/Invoice/RewriteInvoiceRichTextLinksTest.php',
        newPath: 'tests/Billing/Archive/Detailed/Application/RewriteInvoiceRichTextLinksTest.php',
        oldSymbol: 'App\Tests\Application\Billing\Invoice\RewriteInvoiceRichTextLinksTest',
        newSymbol: 'App\Tests\Billing\Archive\Detailed\Application\RewriteInvoiceRichTextLinksTest',
        oldNamespace: 'App\Tests\Application\Billing\Invoice',
        newNamespace: 'App\Tests\Billing\Archive\Detailed\Application',
        shortName: 'RewriteInvoiceRichTextLinksTest',
    );

    $replacements = $rule->collect($context, new AnalysisContext([$mapping->oldSymbol => $mapping]));
    assertSameValue(0, \count($replacements));
});

test('namespace local dependency import rule ignores same namespace function calls', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\NamespaceLocalDependencyImportRule();
    $context = php_context(<<<'PHP'
        <?php

        declare(strict_types=1);

        namespace App\Billing\Domain;

        final class InvoiceBatch
        {
            public function project(): string
            {
                return captureRange();
            }
        }
        PHP, 'src/Billing/Domain/InvoiceBatch.php');

    $replacements = $rule->collect($context, php_analysis_context_for_namespace_local_dependency_move());
    assertSameValue(0, \count($replacements));
});

test('namespace local dependency import rule adds imports for same namespace moved symbols in consumer files', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\NamespaceLocalDependencyImportRule();
    $context = php_context(<<<'PHP'
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
        PHP, 'src/Billing/Domain/InvoiceArchive.php');

    $replacements = $rule->collect($context, php_analysis_context_for_namespace_local_dependency_move());
    assertSameValue(1, \count($replacements));
    assertSameValue(
        "use App\\Billing\\Archive\\Domain\\InvoiceBatch;\n\n",
        $replacements[0]->replacement,
    );
});

test('fully qualified class rule updates exact fqcn references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\FullyQualifiedClassNameReplacementRule();
    $context = php_context("<?php\nreturn new \\App\\Services\\Billing\\InvoiceService();\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('\\App\Domain\Billing\InvoiceService', $replacements[0]->replacement);
});

test('fully qualified class rule preserves imported short style in expressions', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\FullyQualifiedClassNameReplacementRule();
    $context = php_context(<<<'PHP'
        <?php
        use App\Services\Billing\InvoiceService;
        if (!$service instanceof InvoiceService) {
            return new InvoiceService();
        }
        PHP);
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(2, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
    assertSameValue('InvoiceService', $replacements[1]->replacement);
});

test('fully qualified class rule preserves same namespace short style in expressions', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\FullyQualifiedClassNameReplacementRule();
    $context = php_context(<<<'PHP'
        <?php
        namespace App\Billing\Domain;
        final class InvoiceArchive
        {
            public function hasChanges(?InvoiceBatch $changes): bool
            {
                return $changes instanceof InvoiceBatch;
            }
        }
        PHP, 'src/Billing/Domain/InvoiceArchive.php');
    $replacements = $rule->collect($context, php_analysis_context_for_namespace_local_dependency_move());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceBatch', $replacements[0]->replacement);
});

test('class constant rule updates class constant references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\ClassConstantReplacementRule();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\nreturn InvoiceService::class;\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('docblock var rule updates @var references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\DocblockVarReplacementRule();
    $context = php_context("<?php\n/** @var App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($rule->collect($context, php_analysis_context())));
});

test('docblock param rule updates @param references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\DocblockParamReplacementRule();
    $context = php_context(<<<'PHP'
        <?php
        /** @param App\Services\Billing\InvoiceService $service */
        PHP);
    assertSameValue(1, \count($rule->collect($context, php_analysis_context())));
});

test('docblock return rule updates @return references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\DocblockReturnReplacementRule();
    $context = php_context("<?php\n/** @return App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($rule->collect($context, php_analysis_context())));
});

test('docblock throws rule updates @throws references', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\DocblockThrowsReplacementRule();
    $context = php_context("<?php\n/** @throws App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($rule->collect($context, php_analysis_context())));
});

test('attribute class reference rule updates class references inside attributes', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\AttributeClassReferenceReplacementRule();
    $context = php_context("<?php\n#[Attr(service: \\App\\Services\\Billing\\InvoiceService::class)]\nfinal class A {}\n");
    assertSameValue(1, \count($rule->collect($context, php_analysis_context())));
});

test('attribute class reference rule preserves imported short style', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\AttributeClassReferenceReplacementRule();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\n#[Attr(service: InvoiceService::class)]\nfinal class A {}\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('typed property rule updates property types', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\TypedPropertyReplacementRule();
    $context = php_context(<<<'PHP'
        <?php
        use App\Services\Billing\InvoiceService;
        final class A { private InvoiceService $service; }
        PHP);
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('method parameter type rule updates parameter types', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\MethodParameterTypeReplacementRule();
    $context = php_context(<<<'PHP'
        <?php
        use App\Services\Billing\InvoiceService;
        function demo(InvoiceService $service): void {}
        PHP);
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('method return type rule updates return types', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\MethodReturnTypeReplacementRule();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\nfunction demo(): InvoiceService { return new InvoiceService(); }\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('method return type rule preserves fully qualified style even with import present', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\MethodReturnTypeReplacementRule();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\nfunction demo(): \\App\\Services\\Billing\\InvoiceService { return new InvoiceService(); }\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('\\App\\Domain\\Billing\\InvoiceService', $replacements[0]->replacement);
});

test('method return type rule preserves aliased import style', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Php\Rules\MethodReturnTypeReplacementRule();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService as BillingInvoice;\nfunction demo(): BillingInvoice { return new BillingInvoice(); }\n");
    $replacements = $rule->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('BillingInvoice', $replacements[0]->replacement);
});
