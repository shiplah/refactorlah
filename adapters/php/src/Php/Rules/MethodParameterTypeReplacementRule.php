<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Rules;

use PhpParser\Node\Param;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

use function array_merge;

final class MethodParameterTypeReplacementRule extends \Refactorlah\PhpAdapter\Php\Rules\AbstractTypeReplacementRule
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<Param> $parameters */
        $parameters = $finder->findInstanceOf($context->ast, Param::class);

        $replacements = [];
        foreach ($parameters as $parameter) {
            $replacements = array_merge(
                $replacements,
                $this->collectTypeReplacements($context, $analysisContext, $parameter->type, 'php-method-parameter-type')
            );
        }

        return $replacements;
    }
}
