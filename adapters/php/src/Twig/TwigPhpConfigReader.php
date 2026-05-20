<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

use PhpParser\Node;
use PhpParser\Node\Expr\MethodCall;
use PhpParser\NodeFinder;
use PhpParser\Node\Name;
use PhpParser\Node\Scalar\String_;
use PhpParser\ParserFactory;

final class TwigPhpConfigReader
{
    public function read(string $projectRoot): TwigPathConfiguration
    {
        $configPath = $projectRoot . '/config/packages/twig.php';
        if (!is_file($configPath)) {
            return new TwigPathConfiguration([]);
        }

        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $ast = $parser->parse((string) file_get_contents($configPath)) ?? [];
        $finder = new NodeFinder();

        /** @var list<MethodCall> $calls */
        $calls = $finder->findInstanceOf($ast, MethodCall::class);
        $roots = [];

        foreach ($calls as $call) {
            if (!$call->name instanceof Node\Identifier) {
                continue;
            }

            $method = $call->name->toString();
            if ($method === 'defaultPath') {
                $path = $this->extractKernelProjectDirPath($call, 0);
                if ($path !== null) {
                    $roots[] = new TwigPathRoot($path);
                }
                continue;
            }

            if ($method === 'path') {
                $path = $this->extractKernelProjectDirPath($call, 0);
                $namespace = $this->extractStringArgument($call, 1);
                if ($path !== null) {
                    $roots[] = new TwigPathRoot($path, $namespace);
                }
            }
        }

        return new TwigPathConfiguration($roots);
    }

    private function extractKernelProjectDirPath(MethodCall $call, int $index): ?string
    {
        $value = $this->extractStringArgument($call, $index);
        if ($value === null) {
            return null;
        }

        if (!str_starts_with($value, '%kernel.project_dir%/')) {
            return null;
        }

        return trim(substr($value, strlen('%kernel.project_dir%/')), '/');
    }

    private function extractStringArgument(MethodCall $call, int $index): ?string
    {
        if (!isset($call->args[$index])) {
            return null;
        }

        $value = $call->args[$index]->value;
        if ($value instanceof String_) {
            return $value->value;
        }

        return null;
    }
}
