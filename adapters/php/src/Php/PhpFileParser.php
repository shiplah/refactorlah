<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Error;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;

use function array_values;
use function file_get_contents;

final class PhpFileParser
{
    /**
     * @param list<string> $files
     * @return list<PhpFileContext>
     */
    public function parse(string $projectRoot, array $files): array
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $parentConnector = new AstParentConnector();
        $contexts = [];

        foreach ($files as $file) {
            $content = (string) file_get_contents($projectRoot . '/' . $file);
            try {
                $ast = array_values($parser->parse($content) ?? []);
                $traverser = new NodeTraverser();
                $traverser->addVisitor(new NameResolver(options: ['preserveOriginalNames' => true]));
                $resolved = array_values($traverser->traverse($ast));
                $parentConnector->attach($resolved);
                $contexts[] = new PhpFileContext($file, $content, $resolved);
            } catch (Error) {
                continue;
            }
        }

        return $contexts;
    }
}
