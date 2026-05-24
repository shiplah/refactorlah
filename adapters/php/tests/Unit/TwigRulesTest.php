<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\Config\PathMapping;
use Refactorlah\PhpAdapter\Protocol\MoveCollection;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigConfigReader;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigPathConfiguration;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigPathRoot;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigTemplateMapper;

function twig_mapping(): PathMapping
{
    return new PathMapping(
        kind: 'twig-template',
        oldPath: 'templates/admin/user/card.html.twig',
        newPath: 'templates/backoffice/user/card.html.twig',
        oldReference: 'admin/user/card.html.twig',
        newReference: 'backoffice/user/card.html.twig',
    );
}

function twig_directory_mapping(): PathMapping
{
    return new PathMapping(
        kind: 'twig-template-directory',
        oldPath: 'src/Billing/FileTree/Ui/Web/Twig/file-tree.html.twig',
        newPath: 'src/Billing/Reminder/Ui/Web/Twig/file-tree.html.twig',
        oldReference: '@Billing/FileTree/Ui/Web/Twig',
        newReference: '@Billing/Reminder/Ui/Web/Twig',
    );
}

test('twig template mapper derives deterministic template references', function (): void
{
    $mappings = (new TwigTemplateMapper())->deriveMappings(MoveCollection::fromMixed([[
        'oldPath' => 'templates/admin/user/card.html.twig',
        'newPath' => 'templates/backoffice/user/card.html.twig',
        'tracked' => true,
    ]]), new TwigPathConfiguration([new TwigPathRoot('templates')]));

    assertSameValue(2, \count($mappings));
    assertSameValue('admin/user/card.html.twig', $mappings[0]->oldReference);
    assertSameValue('admin/user', $mappings[1]->oldReference);
});

test('twig template mapper derives alias references from configured twig paths', function (): void
{
    $mappings = (new TwigTemplateMapper())->deriveMappings(MoveCollection::fromMixed([[
        'oldPath' => 'templates/billing/archive.html.twig',
        'newPath' => 'src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig',
        'tracked' => true,
    ]]), new TwigPathConfiguration([
        new TwigPathRoot('templates'),
        new TwigPathRoot('src/Billing', 'Billing'),
    ]));

    assertSameValue(2, \count($mappings));
    assertSameValue('billing/archive.html.twig', $mappings[0]->oldReference);
    assertSameValue('@Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig', $mappings[0]->newReference);
    assertSameValue('billing', $mappings[1]->oldReference);
    assertSameValue('@Billing/Archive/Listing/Ui/Web/Twig', $mappings[1]->newReference);
});

test('twig include rule updates include statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigIncludeReplacementRule();
    $replacements = $rule->collect('templates/demo.html.twig', "{% include 'admin/user/card.html.twig' %}", twig_mapping());
    assertSameValue(1, \count($replacements));
});

test('twig extends rule updates extends statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigExtendsReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% extends 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig embed rule updates embed statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigEmbedReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% embed 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig use rule updates use statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigUseReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% use 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig import rule updates import statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigImportReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% import 'admin/user/card.html.twig' as macros %}", twig_mapping())));
});

test('twig from rule updates from statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TwigFromReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% from 'admin/user/card.html.twig' import badge %}", twig_mapping())));
});

test('symfony render rule updates render template strings', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\RenderTemplateReplacementRule();
    assertSameValue(1, \count($rule->collect('app/Controller.php', "<?php \$this->render('admin/user/card.html.twig');", twig_mapping())));
});

test('symfony template attribute rule updates attribute template strings', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\TemplateAttributeReplacementRule();
    assertSameValue(1, \count($rule->collect('app/Controller.php', "<?php #[Template('admin/user/card.html.twig')]", twig_mapping())));
});

test('twig component template attribute rule updates template strings', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\ComponentTemplateAttributeReplacementRule();
    $replacements = $rule->collect(
        'src/Component.php',
        "<?php #[AsTwigComponent(template: '@Billing/FileTree/Ui/Web/Twig/file-tree.html.twig')]",
        new PathMapping(
            kind: 'twig-template',
            oldPath: twig_directory_mapping()->oldPath,
            newPath: twig_directory_mapping()->newPath,
            oldReference: '@Billing/FileTree/Ui/Web/Twig/file-tree.html.twig',
            newReference: '@Billing/Reminder/Ui/Web/Twig/file-tree.html.twig',
        ),
    );
    assertSameValue(1, \count($replacements));
    assertSameValue("'@Billing/Reminder/Ui/Web/Twig/file-tree.html.twig'", $replacements[0]->replacement);
});

test('yaml twig template rule updates template fields', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\YamlTemplateReplacementRule();
    assertSameValue(1, \count($rule->collect('config/routes.yaml', "template: 'admin/user/card.html.twig'\n", twig_mapping())));
});

test('yaml twig component template directory rule updates template directories', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Symfony\Twig\Rules\YamlComponentTemplateDirectoryReplacementRule();
    $replacements = $rule->collect(
        'config/packages/twig_component.yaml',
        "template_directory: '@Billing/FileTree/Ui/Web/Twig'\n",
        twig_directory_mapping(),
    );
    assertSameValue(1, \count($replacements));
    assertSameValue("'@Billing/Reminder/Ui/Web/Twig'", $replacements[0]->replacement);
});

test('yaml path scanner updates exact asset mapper paths', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-yaml-path-' . \uniqid();
    \mkdir($root . '/config/packages', 0o777, true);
    \file_put_contents($root . '/config/packages/asset_mapper.yaml', <<<'YAML'
        framework:
          asset_mapper:
            paths:
              - 'src/Shared/Ui/Web/'
              - 'assets/'
        YAML);

    $replacements = (new \Refactorlah\PhpAdapter\Symfony\Core\YamlAssetMapperPathReferenceScanner())->scan(
        projectRoot: $root,
        files: ['config/packages/asset_mapper.yaml'],
        pathMappings: (new \Refactorlah\PhpAdapter\Config\PathMappingFactory())->fromMove(
            'src/Shared/Ui/Web',
            'src/Shared/Controls/Ui/Web',
        ),
    );

    assertSameValue(1, \count($replacements));
    assertSameValue("'src/Shared/Controls/Ui/Web/'", $replacements[0]->replacement);
});

test('yaml path scanner skips non asset mapper path strings', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-yaml-path-' . \uniqid();
    \mkdir($root . '/config/packages', 0o777, true);
    \file_put_contents($root . '/config/packages/example.yaml', <<<'YAML'
        something_else:
          paths:
            - 'src/Shared/Ui/Web/'
        YAML);

    $replacements = (new \Refactorlah\PhpAdapter\Symfony\Core\YamlAssetMapperPathReferenceScanner())->scan(
        projectRoot: $root,
        files: ['config/packages/example.yaml'],
        pathMappings: (new \Refactorlah\PhpAdapter\Config\PathMappingFactory())->fromMove(
            'src/Shared/Ui/Web',
            'src/Shared/Controls/Ui/Web',
        ),
    );

    assertSameValue(0, \count($replacements));
});

test('static import scanner updates exact moved asset imports', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-static-import-' . \uniqid();
    \mkdir($root . '/assets', 0o777, true);

    $content = "import '../src/Billing/Archive/Listing/Ui/Web/Twig/invoice-line-preview.css';\n";
    \file_put_contents($root . '/assets/app.js', $content);

    $replacements = (new \Refactorlah\PhpAdapter\Config\StaticImportReferenceScanner())->scan(
        projectRoot: $root,
        files: ['assets/app.js'],
        moves: MoveCollection::fromMixed([[
            'oldPath' => 'src/Billing/Archive/Listing/Ui/Web/Twig/invoice-line-preview.css',
            'newPath' => 'src/Billing/Archive/InvoiceLinePreview/Ui/Web/Twig/invoice-line-preview.css',
            'tracked' => true,
        ]]),
    );

    assertSameValue(1, \count($replacements));
    assertSameValue('../src/Billing/Archive/InvoiceLinePreview/Ui/Web/Twig/invoice-line-preview.css', $replacements[0]->replacement);
});

test('twig registry warns on dynamic template paths', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-warning-' . \uniqid();
    \mkdir($root . '/app', 0o777, true);
    \file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template ?: 'admin/user/card.html.twig');\n");

    [$replacements, $warnings] = (new \Refactorlah\PhpAdapter\Symfony\Twig\TwigRuleRegistry())->scan(
        projectRoot: $root,
        files: ['app/Controller.php'],
        twigFiles: [],
        pathMappings: [twig_mapping()],
    );

    assertSameValue(0, \count($replacements));
    assertTrueValue(\count($warnings) >= 1, 'expected at least one warning');
});

test('twig config reader supports php-based symfony twig config', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-config-' . \uniqid();
    \mkdir($root . '/config/packages', 0o777, true);
    \file_put_contents($root . '/config/packages/twig.php', <<<'PHP'
        <?php

        use Symfony\Config\TwigConfig;

        return static function (TwigConfig $twig): void {
            $twig->defaultPath('%kernel.project_dir%/templates');
            $twig->path('%kernel.project_dir%/src/Billing', 'Billing');
        };
        PHP);

    $config = (new TwigConfigReader())->read($root);
    assertSameValue(2, \count($config->roots));
    assertSameValue('templates', $config->roots[0]->path);
    assertSameValue('src/Billing', $config->roots[1]->path);
    assertSameValue('Billing', $config->roots[1]->namespace);
});

test('twig config reader supports yaml path aliases', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-config-' . \uniqid();
    \mkdir($root . '/config/packages', 0o777, true);
    \file_put_contents($root . '/config/packages/twig.yaml', <<<'YAML'
        twig:
          default_path: '%kernel.project_dir%/templates'
          paths:
            '%kernel.project_dir%/src/Billing': Billing
        YAML);

    $config = (new TwigConfigReader())->read($root);
    assertSameValue(2, \count($config->roots));
    assertSameValue('templates', $config->roots[0]->path);
    assertSameValue('src/Billing', $config->roots[1]->path);
    assertSameValue('Billing', $config->roots[1]->namespace);
});

test('twig template mapper prefers the longest matching root', function (): void
{
    $mappings = (new TwigTemplateMapper())->deriveMappings(MoveCollection::fromMixed([[
        'oldPath' => 'src/Billing/Archive/card.html.twig',
        'newPath' => 'src/Billing/Archive/Listing/card.html.twig',
        'tracked' => true,
    ]]), new TwigPathConfiguration([
        new TwigPathRoot('src/Billing', 'Billing'),
        new TwigPathRoot('src/Billing/Archive', 'Archive'),
    ]));

    assertSameValue(2, \count($mappings));
    assertSameValue('@Archive/card.html.twig', $mappings[0]->oldReference);
    assertSameValue('@Archive/Listing/card.html.twig', $mappings[0]->newReference);
});

test('twig registry does not warn on unrelated dynamic render variables', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-dynamic-' . \uniqid();
    \mkdir($root . '/app', 0o777, true);
    \file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template);\n");

    [$replacements, $warnings] = (new \Refactorlah\PhpAdapter\Symfony\Twig\TwigRuleRegistry())->scan(
        projectRoot: $root,
        files: ['app/Controller.php'],
        twigFiles: [],
        pathMappings: [twig_mapping()],
    );

    assertSameValue(0, \count($replacements));
    assertSameValue(0, \count($warnings));
});
