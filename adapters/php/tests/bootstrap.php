<?php

declare(strict_types=1);

require \dirname(__DIR__) . '/vendor/autoload.php';

/** @var list<array{0:string,1:Closure():void}> $__tests */
$__tests = [];

function test(string $name, Closure $closure): void
{
    /** @var list<array{0:string,1:Closure():void}> $__tests */
    global $__tests;
    $__tests[] = [$name, $closure];
}

function assertSameValue(mixed $expected, mixed $actual, string $message = ''): void
{
    if ($expected !== $actual) {
        throw new RuntimeException('' !== $message ? $message : \sprintf('Expected %s, got %s', \var_export($expected, true), \var_export($actual, true)));
    }
}

function assertTrueValue(bool $condition, string $message): void
{
    if (!$condition) {
        throw new RuntimeException($message);
    }
}

function run_all_tests(): int
{
    /** @var list<array{0:string,1:Closure():void}> $__tests */
    global $__tests;

    $failures = 0;
    foreach ($__tests as [$name, $closure]) {
        try {
            $closure();
            echo "PASS {$name}\n";
        } catch (Throwable $throwable) {
            $failures++;
            \fwrite(STDERR, "FAIL {$name}: {$throwable->getMessage()}\n");
        }
    }

    return 0 === $failures ? 0 : 1;
}
