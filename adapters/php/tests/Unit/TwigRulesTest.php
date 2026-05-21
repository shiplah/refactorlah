<?php

declare(strict_types=1);

use Refactorlah\PhpAdapter\Twig\TwigConfigReader;
use Refactorlah\PhpAdapter\Twig\TwigPathConfiguration;
use Refactorlah\PhpAdapter\Twig\TwigPathRoot;
use Refactorlah\PhpAdapter\Twig\TwigTemplateMapper;

/**
 * @return array{
 *   kind:string,
 *   oldPath:string,
 *   newPath:string,
 *   oldReference:string,
 *   newReference:string
 * }
 */
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

test('twig template mapper derives deterministic template references', function (): void
{
    $mappings = (new TwigTemplateMapper())->deriveMappings([[
        'oldPath' => 'templates/admin/user/card.html.twig',
        'newPath' => 'templates/backoffice/user/card.html.twig',
        'tracked' => true,
    ]], new TwigPathConfiguration([new TwigPathRoot('templates')]));

    assertSameValue(1, \count($mappings));
    assertSameValue('admin/user/card.html.twig', $mappings[0]['oldReference']);
});

test('twig template mapper derives alias references from configured twig paths', function (): void
{
    $mappings = (new TwigTemplateMapper())->deriveMappings([[
        'oldPath' => 'templates/billing/archive.html.twig',
        'newPath' => 'src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig',
        'tracked' => true,
    ]], new TwigPathConfiguration([
        new TwigPathRoot('templates'),
        new TwigPathRoot('src/Billing', 'Billing'),
    ]));

    assertSameValue(1, \count($mappings));
    assertSameValue('billing/archive.html.twig', $mappings[0]['oldReference']);
    assertSameValue('@Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig', $mappings[0]['newReference']);
});

test('twig include rule updates include statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigIncludeReplacementRule();
    $replacements = $rule->collect('templates/demo.html.twig', "{% include 'admin/user/card.html.twig' %}", twig_mapping());
    assertSameValue(1, \count($replacements));
});

test('twig extends rule updates extends statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigExtendsReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% extends 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig embed rule updates embed statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigEmbedReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% embed 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig use rule updates use statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigUseReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% use 'admin/user/card.html.twig' %}", twig_mapping())));
});

test('twig import rule updates import statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigImportReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% import 'admin/user/card.html.twig' as macros %}", twig_mapping())));
});

test('twig from rule updates from statements', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\TwigFromReplacementRule();
    assertSameValue(1, \count($rule->collect('templates/demo.html.twig', "{% from 'admin/user/card.html.twig' import badge %}", twig_mapping())));
});

test('symfony render rule updates render template strings', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\SymfonyRenderTemplateReplacementRule();
    assertSameValue(1, \count($rule->collect('app/Controller.php', "<?php \$this->render('admin/user/card.html.twig');", twig_mapping())));
});

test('symfony template attribute rule updates attribute template strings', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\SymfonyTemplateAttributeReplacementRule();
    assertSameValue(1, \count($rule->collect('app/Controller.php', "<?php #[Template('admin/user/card.html.twig')]", twig_mapping())));
});

test('yaml twig template rule updates template fields', function (): void
{
    $rule = new \Refactorlah\PhpAdapter\Twig\Rules\YamlTwigTemplateReplacementRule();
    assertSameValue(1, \count($rule->collect('config/routes.yaml', "template: 'admin/user/card.html.twig'\n", twig_mapping())));
});

test('twig registry warns on dynamic template paths', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-warning-' . \uniqid();
    \mkdir($root . '/app', 0o777, true);
    \file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template ?: 'admin/user/card.html.twig');\n");

    [$replacements, $warnings] = (new \Refactorlah\PhpAdapter\Twig\TwigRuleRegistry())->scan(
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
    $mappings = (new TwigTemplateMapper())->deriveMappings([[
        'oldPath' => 'src/Billing/Archive/card.html.twig',
        'newPath' => 'src/Billing/Archive/Listing/card.html.twig',
        'tracked' => true,
    ]], new TwigPathConfiguration([
        new TwigPathRoot('src/Billing', 'Billing'),
        new TwigPathRoot('src/Billing/Archive', 'Archive'),
    ]));

    assertSameValue(1, \count($mappings));
    assertSameValue('@Archive/card.html.twig', $mappings[0]['oldReference']);
    assertSameValue('@Archive/Listing/card.html.twig', $mappings[0]['newReference']);
});

test('twig registry does not warn on unrelated dynamic render variables', function (): void
{
    $root = \sys_get_temp_dir() . '/refactorlah-twig-dynamic-' . \uniqid();
    \mkdir($root . '/app', 0o777, true);
    \file_put_contents($root . '/app/Controller.php', "<?php \$this->render(\$template);\n");

    [$replacements, $warnings] = (new \Refactorlah\PhpAdapter\Twig\TwigRuleRegistry())->scan(
        projectRoot: $root,
        files: ['app/Controller.php'],
        twigFiles: [],
        pathMappings: [twig_mapping()],
    );

    assertSameValue(0, \count($replacements));
    assertSameValue(0, \count($warnings));
});
