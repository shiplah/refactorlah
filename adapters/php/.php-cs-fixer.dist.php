<?php

declare(strict_types=1);

use PhpCsFixer\Runner\Parallel\ParallelConfigFactory;

$finder = PhpCsFixer\Finder::create()
    ->in(__DIR__ . '/src')
    ->in(__DIR__ . '/tests')
    ->append([
        __DIR__ . '/bin/refactorlah-php',
    ])
;

return new PhpCsFixer\Config()
    ->setParallelConfig(ParallelConfigFactory::detect())
    ->setRiskyAllowed(true)
    ->setRules([
        '@PHP82Migration' => true,
        '@PER-CS3x0' => true,
        '@Symfony' => true,
        'array_push' => true,
        'array_syntax' => ['syntax' => 'short'],
        'assign_null_coalescing_to_coalesce_equal' => true,
        'declare_strict_types' => true,
        'final_class' => true,
        'global_namespace_import' => [
            'import_classes' => false,
            'import_constants' => false,
            'import_functions' => true,
        ],
        'long_to_shorthand_operator' => true,
        'mb_str_functions' => true,
        'modernize_types_casting' => true,
        'native_function_invocation' => [
            'include' => ['@internal'],
            'scope' => 'all',
            'strict' => true,
        ],
        'ordered_imports' => [
            'case_sensitive' => true,
            'imports_order' => [
                'class',
                'function',
                'const',
            ],
        ],
        'ordered_types' => [
            'case_sensitive' => true,
            'sort_algorithm' => 'none',
            'null_adjustment' => 'always_last',
        ],
        'phpdoc_line_span' => [
            'const' => 'single',
            'method' => 'single',
            'property' => 'single',
        ],
        'phpdoc_to_comment' => false,
        'single_line_empty_body' => true,
        'strict_comparison' => true,
        'strict_param' => true,
        'ternary_to_null_coalescing' => true,
        'yoda_style' => true,
    ])
    ->setCacheFile(__DIR__ . '/var/fixer/.php-cs-fixer.cache')
    ->setFinder($finder)
    ->setUnsupportedPhpVersionAllowed(true)
;
