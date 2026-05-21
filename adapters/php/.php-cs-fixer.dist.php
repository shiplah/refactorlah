<?php

declare(strict_types=1);

use PhpCsFixer\Runner\Parallel\ParallelConfigFactory;

$rules = [
    '@PER-CS3x0' => true,
    '@PHP82Migration' => true,
    'array_push' => true,
    'assign_null_coalescing_to_coalesce_equal' => true,
    'braces_position' => [
        'anonymous_classes_opening_brace' => 'next_line_unless_newline_at_signature_end',
        'anonymous_functions_opening_brace' => 'next_line_unless_newline_at_signature_end',
    ],
    'declare_strict_types' => true,
    'explicit_string_variable' => false,
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
    'no_alternative_syntax' => true,
    'no_empty_comment' => true,
    'no_empty_phpdoc' => true,
    'no_homoglyph_names' => true,
    'no_multiline_whitespace_around_double_arrow' => true,
    'no_unused_imports' => true,
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
    'strict_comparison' => true,
    'strict_param' => true,
    'ternary_to_null_coalescing' => true,
    'unary_operator_spaces' => true,
    'whitespace_after_comma_in_array' => true,
    'yoda_style' => true,
];

$finder = PhpCsFixer\Finder::create()
    ->in(__DIR__ . '/src')
    ->in(__DIR__ . '/tests')
    ->notPath('fixtures')
    ->append([
        __DIR__ . '/bin/refactorlah-php',
    ]);

return new PhpCsFixer\Config()
    ->setParallelConfig(ParallelConfigFactory::detect())
    ->setRiskyAllowed(true)
    ->setRules($rules)
    ->setCacheFile(__DIR__ . '/var/fixer/.php-cs-fixer.cache')
    ->setFinder($finder)
    ->setUnsupportedPhpVersionAllowed(true)
;
