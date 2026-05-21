<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php\Workers;

use PhpParser\Node\Stmt\Property;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpFileContext;

use function array_merge;

final class TypedPropertyReplacementWorker extends AbstractTypeReplacementWorker
{
    public function name(): string
    {
        return self::class;
    }

    public function collect(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        /** @var list<Property> $properties */
        $properties = $finder->findInstanceOf($context->ast, Property::class);

        $replacements = [];
        foreach ($properties as $property) {
            $replacements = array_merge(
                $replacements,
                $this->collectTypeReplacements($context, $analysisContext, $property->type, 'php-typed-property')
            );
        }

        return $replacements;
    }
}
