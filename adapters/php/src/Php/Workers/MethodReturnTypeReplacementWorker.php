<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Expr\ArrowFunction;
use PhpParser\Node\Expr\Closure;
use PhpParser\Node\FunctionLike;
use PhpParser\Node\Stmt\ClassMethod;
use PhpParser\Node\Stmt\Function_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

final class MethodReturnTypeReplacementWorker extends AbstractTypeReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        $functionLikes = array_merge(
            $finder->findInstanceOf($context->ast, ClassMethod::class),
            $finder->findInstanceOf($context->ast, Function_::class),
            $finder->findInstanceOf($context->ast, Closure::class),
            $finder->findInstanceOf($context->ast, ArrowFunction::class),
        );

        $replacements = [];
        foreach ($functionLikes as $functionLike) {
            if (!$functionLike instanceof FunctionLike) {
                continue;
            }
            $replacements = array_merge(
                $replacements,
                $this->collectTypeReplacements($context, $analysisContext, $functionLike->getReturnType(), 'php-method-return-type')
            );
        }

        return $replacements;
    }
}
