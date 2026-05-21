<?php

declare(strict_types=1);

use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\SymbolMapping;
use Refactorlah\PhpAdapter\Php\WorkerSupport;
use Refactorlah\PhpAdapter\Php\Workers\AttributeClassReferenceReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\ClassConstantReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockParamReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockReturnReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockThrowsReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\DocblockVarReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\FullyQualifiedClassNameReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\GroupUseStatementReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\MethodParameterTypeReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\MethodReturnTypeReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\NamespaceDeclarationReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\TypedPropertyReplacementWorker;
use Refactorlah\PhpAdapter\Php\Workers\UseStatementReplacementWorker;

function php_context(string $content, string $path = 'app/Http/Controllers/InvoiceController.php'): PhpFileContext
{
    $parser = (new ParserFactory())->createForNewestSupportedVersion();
    $ast = $parser->parse($content) ?? [];
    $traverser = new NodeTraverser();
    $traverser->addVisitor(new NameResolver(options: ['preserveOriginalNames' => true]));
    $resolved = $traverser->traverse($ast);
    WorkerSupport::attachParents($resolved);
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

test('namespace declaration worker updates moved file namespace', function (): void
{
    $worker = new NamespaceDeclarationReplacementWorker();
    $context = php_context("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n", 'app/Services/Billing/InvoiceService.php');
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('App\Domain\Billing', $replacements[0]->replacement);
});

test('use statement worker updates imported symbol', function (): void
{
    $worker = new UseStatementReplacementWorker();
    $context = php_context("<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n");
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
});

test('group use worker skips conservatively', function (): void
{
    $worker = new GroupUseStatementReplacementWorker();
    $context = php_context("<?php\nuse App\\Services\\Billing\\{InvoiceService};\n");
    assertSameValue(0, \count($worker->collect($context, php_analysis_context())));
});

test('fully qualified class worker updates exact fqcn references', function (): void
{
    $worker = new FullyQualifiedClassNameReplacementWorker();
    $context = php_context("<?php\nreturn new \\App\\Services\\Billing\\InvoiceService();\n");
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('\\App\Domain\Billing\InvoiceService', $replacements[0]->replacement);
});

test('class constant worker updates class constant references', function (): void
{
    $worker = new ClassConstantReplacementWorker();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\nreturn InvoiceService::class;\n");
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
});

test('docblock var worker updates @var references', function (): void
{
    $worker = new DocblockVarReplacementWorker();
    $context = php_context("<?php\n/** @var App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($worker->collect($context, php_analysis_context())));
});

test('docblock param worker updates @param references', function (): void
{
    $worker = new DocblockParamReplacementWorker();
    $context = php_context(<<<'PHP'
        <?php
        /** @param App\Services\Billing\InvoiceService $service */
        PHP);
    assertSameValue(1, \count($worker->collect($context, php_analysis_context())));
});

test('docblock return worker updates @return references', function (): void
{
    $worker = new DocblockReturnReplacementWorker();
    $context = php_context("<?php\n/** @return App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($worker->collect($context, php_analysis_context())));
});

test('docblock throws worker updates @throws references', function (): void
{
    $worker = new DocblockThrowsReplacementWorker();
    $context = php_context("<?php\n/** @throws App\\Services\\Billing\\InvoiceService */\n");
    assertSameValue(1, \count($worker->collect($context, php_analysis_context())));
});

test('attribute class reference worker updates class references inside attributes', function (): void
{
    $worker = new AttributeClassReferenceReplacementWorker();
    $context = php_context("<?php\n#[Attr(service: \\App\\Services\\Billing\\InvoiceService::class)]\nfinal class A {}\n");
    assertSameValue(1, \count($worker->collect($context, php_analysis_context())));
});

test('typed property worker updates property types', function (): void
{
    $worker = new TypedPropertyReplacementWorker();
    $context = php_context(<<<'PHP'
        <?php
        use App\Services\Billing\InvoiceService;
        final class A { private InvoiceService $service; }
        PHP);
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('method parameter type worker updates parameter types', function (): void
{
    $worker = new MethodParameterTypeReplacementWorker();
    $context = php_context(<<<'PHP'
        <?php
        use App\Services\Billing\InvoiceService;
        function demo(InvoiceService $service): void {}
        PHP);
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});

test('method return type worker updates return types', function (): void
{
    $worker = new MethodReturnTypeReplacementWorker();
    $context = php_context("<?php\nuse App\\Services\\Billing\\InvoiceService;\nfunction demo(): InvoiceService { return new InvoiceService(); }\n");
    $replacements = $worker->collect($context, php_analysis_context());
    assertSameValue(1, \count($replacements));
    assertSameValue('InvoiceService', $replacements[0]->replacement);
});
