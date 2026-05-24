<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Symfony\Twig;

use PhpParser\Node;
use PhpParser\Node\Arg;
use PhpParser\Node\Expr\MethodCall;
use PhpParser\Node\Scalar\String_;
use PhpParser\NodeFinder;
use PhpParser\ParserFactory;

use function file_get_contents;
use function is_file;
use function mb_strlen;
use function mb_substr;
use function mb_trim;
use function str_starts_with;

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
            if ('defaultPath' === $method) {
                $path = $this->extractKernelProjectDirPath($call, 0);
                if (null !== $path) {
                    $roots[] = new TwigPathRoot($path);
                }
                continue;
            }

            if ('path' === $method) {
                $path = $this->extractKernelProjectDirPath($call, 0);
                $namespace = $this->extractStringArgument($call, 1);
                if (null !== $path) {
                    $roots[] = new TwigPathRoot($path, $namespace);
                }
            }
        }

        return new TwigPathConfiguration($roots);
    }

    private function extractKernelProjectDirPath(MethodCall $call, int $index): ?string
    {
        $value = $this->extractStringArgument($call, $index);
        if (null === $value) {
            return null;
        }

        if (!str_starts_with($value, '%kernel.project_dir%/')) {
            return null;
        }

        return mb_trim(mb_substr($value, mb_strlen('%kernel.project_dir%/')), '/');
    }

    private function extractStringArgument(MethodCall $call, int $index): ?string
    {
        if (!isset($call->args[$index])) {
            return null;
        }

        $argument = $call->args[$index];
        if (!$argument instanceof Arg) {
            return null;
        }

        $value = $argument->value;
        if ($value instanceof String_) {
            return $value->value;
        }

        return null;
    }
}
