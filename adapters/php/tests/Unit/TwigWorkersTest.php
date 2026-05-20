<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\Twig\TwigTemplateMapper;
use Refactorlah\PhpAdapter\Twig\TwigConfigReader;
use Refactorlah\PhpAdapter\Twig\TwigPathConfiguration;
use Refactorlah\PhpAdapter\Twig\TwigPathRoot;
use Refactorlah\PhpAdapter\Twig\TwigWorkerRegistry;
use Refactorlah\PhpAdapter\Twig\Workers\SymfonyRenderTemplateReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\SymfonyTemplateAttributeReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigEmbedReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigExtendsReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigFromReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigImportReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigIncludeReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigUseReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\YamlTwigTemplateReplacementWorker;

function twig_mapping(): array
{
    return [
        'kind' => 'twig-template',
        'oldPath' => 'templates/admin/user/card.html.twig',
        'newPath' => 'templates/backoffice/user/card.html.twig',
        'oldReference' => 'admin/user/card.html.twig',
        'newReference' => 'backoffice/user/card.html.twig',
    ];
}

test('twig template mapper derives deterministic template references', function (): void {
    $mappings = (new TwigTemplateMapper())->deriveMappings([[
        'oldPath' => 'templates/admin/user/card.html.twig',
        'newPath' => 'templates/backoffice/user/card.html.twig',
        'tracked' => true,
    ]], new TwigPathConfiguration([new TwigPathRoot('templates')]));

    assertSameValue(1, count($mappings));
    assertSameValue('admin/user/card.html.twig', $mappings[0]['oldReference']);
});

test('twig template mapper derives alias references from configured twig paths', function (): void {
    $mappings = (new TwigTemplateMapper())->deriveMappings([[
        'oldPath' => 'templates/billing/archive.html.twig',
        'newPath' => 'src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig',
        'tracked' => true,
    ]], new TwigPathConfiguration([
        new TwigPathRoot('templates'),
        new TwigPathRoot('src/Billing', 'Billing'),
    ]));

    assertSameValue(1, count($mappings));
    assertSameValue('billing/archive.html.twig', $mappings[0]['oldReference']);
    assertSameValue('@Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig', $mappings[0]['newReference']);
});

test('twig include worker updates include statements', function (): void {
    $worker = new TwigIncludeReplacementWorker();
    $replacements = $worker->collect('templates/demo.html.twig', "{% include 'admin/user/card.html.twig' %}", twig_mapping());
    assertSameValue(1, count($replacements));
});

test('twig extends worker updates extends statements', function (): void {
    $worker = new TwigExtendsReplacementWorker();
    assertSameValue(1, count($worker->collect('templates/demo.html.twig', "{% extends 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig embed worker updates embed statements', function (): void {
    $worker = new TwigEmbedReplacementWorker();
    assertSameValue(1, count($worker->collect('templates/demo.html.twig', "{% embed 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig use worker updates use statements', function (): void {
    $worker = new TwigUseReplacementWorker();
    assertSameValue(1, count($worker->collect('templates/demo.html.twig', "{% use 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig import worker updates import statements', function (): void {
    $worker = new TwigImportReplacementWorker();
    assertSameValue(1, count($worker->collect('templates/demo.html.twig', "{% import 'admin/user/card.html.twig' as macros %}", twig_mapping())));
});

test('twig from worker updates from statements', function (): void {
    $worker = new TwigFromReplacementWorker();
    assertSameValue(1, count($worker->collect('templates/demo.html.twig', "{% from 'admin/user/card.html.twig' import badge %}", twig_mapping())));
});

test('symfony render worker updates render template strings', function (): void {
    $worker = new SymfonyRenderTemplateReplacementWorker();
    assertSameValue(1, count($worker->collect('app/Controller.php', "<?php \$this->render('admin/user/card.html.twig');", twig_mapping())));
});

test('symfony template attribute worker updates attribute template strings', function (): void {
    $worker = new SymfonyTemplateAttributeReplacementWorker();
    assertSameValue(1, count($worker->collect('app/Controller.php', "<?php #[Template('admin/user/card.html.twig')]", twig_mapping())));
});

test('yaml twig template worker updates template fields', function (): void {
    $worker = new YamlTwigTemplateReplacementWorker();
    assertSameValue(1, count($worker->collect('config/routes.yaml', "template: 'admin/user/card.html.twig'\n", twig_mapping())));
});

test('twig registry warns on dynamic template paths', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-twig-warning-' . uniqid();
    mkdir($root . '/app', 0777, true);
    file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template ?: 'admin/user/card.html.twig');\n");

    [$replacements, $warnings] = (new TwigWorkerRegistry())->scan(
        projectRoot: $root,
        files: ['app/Controller.php'],
        twigFiles: [],
        pathMappings: [twig_mapping()],
    );

    assertSameValue(0, count($replacements));
    assertTrueValue(count($warnings) >= 1, 'expected at least one warning');
});

test('twig config reader supports php-based symfony twig config', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-twig-config-' . uniqid();
    mkdir($root . '/config/packages', 0777, true);
    file_put_contents($root . '/config/packages/twig.php', <<<'PHP'
<?php

use Symfony\Config\TwigConfig;

return static function (TwigConfig $twig): void {
    $twig->defaultPath('%kernel.project_dir%/templates');
    $twig->path('%kernel.project_dir%/src/Billing', 'Billing');
};
PHP);

    $config = (new TwigConfigReader())->read($root);
    assertSameValue(2, count($config->roots));
    assertSameValue('templates', $config->roots[0]->path);
    assertSameValue('src/Billing', $config->roots[1]->path);
    assertSameValue('Billing', $config->roots[1]->namespace);
});

test('twig registry does not warn on unrelated dynamic render variables', function (): void {
    $root = sys_get_temp_dir() . '/refactorlah-twig-dynamic-' . uniqid();
    mkdir($root . '/app', 0777, true);
    file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template);\n");

    [$replacements, $warnings] = (new TwigWorkerRegistry())->scan(
        projectRoot: $root,
        files: ['app/Controller.php'],
        twigFiles: [],
        pathMappings: [twig_mapping()],
    );

    assertSameValue(0, count($replacements));
    assertSameValue(0, count($warnings));
});
