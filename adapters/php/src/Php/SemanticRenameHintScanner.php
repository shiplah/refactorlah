<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Php;

use PhpParser\Node;
use PhpParser\Node\Expr\Variable;
use PhpParser\Node\Name;
use PhpParser\Node\Scalar\String_;
use PhpParser\NodeFinder;
use Refactorlah\PhpAdapter\Warning\Warning;

use function file_get_contents;
use function is_string;
use function mb_strpos;
use function mb_substr;
use function mb_substr_count;
use function str_contains;
use function str_replace;

final class SemanticRenameHintScanner
{
    public function __construct(
        private readonly SemanticNameVariants $variants = new SemanticNameVariants(),
    ) {}

    /**
     * @param list<PhpFileContext> $contexts
     * @return list<Warning>
     */
    public function scanPhpContexts(array $contexts, AnalysisContext $analysisContext): array
    {
        $warnings = [];

        foreach ($contexts as $context) {
            $warnings = [
                ...$warnings,
                ...$this->scanPhpContext($context, $analysisContext),
            ];
        }

        return $warnings;
    }

    /**
     * @param list<string> $files
     * @param list<SymbolMapping> $symbolMappings
     * @return list<Warning>
     */
    public function scanTextFiles(string $projectRoot, array $files, array $symbolMappings): array
    {
        $warnings = [];
        foreach ($files as $file) {
            $content = file_get_contents($projectRoot . '/' . $file);
            if (!is_string($content) || '' === $content) {
                continue;
            }

            foreach ($symbolMappings as $mapping) {
                foreach ($this->variants->literalHints($mapping) as $old => $new) {
                    if (!str_contains($content, $old)) {
                        continue;
                    }

                    $offset = mb_strpos($content, $old);
                    $warnings[] = new Warning(
                        message: $this->message($old, $new),
                        file: $file,
                        line: $this->lineForOffset($content, false === $offset ? 0 : $offset),
                    );
                }
            }
        }

        return $warnings;
    }

    /** @return list<Warning> */
    private function scanPhpContext(PhpFileContext $context, AnalysisContext $analysisContext): array
    {
        $finder = new NodeFinder();
        $warnings = [];

        /** @var list<Variable> $variables */
        $variables = $finder->findInstanceOf($context->ast, Variable::class);
        foreach ($variables as $variable) {
            if (!is_string($variable->name)) {
                continue;
            }

            foreach ($analysisContext->symbolMappings as $mapping) {
                foreach ($this->variants->variableHints($mapping) as $old => $new) {
                    if ($variable->name !== $old) {
                        continue;
                    }

                    $warnings[] = new Warning(
                        message: $this->message($old, $new),
                        file: $context->path,
                        line: $variable->getStartLine(),
                    );
                }
            }
        }

        /** @var list<String_> $strings */
        $strings = $finder->findInstanceOf($context->ast, String_::class);
        foreach ($strings as $string) {
            foreach ($analysisContext->symbolMappings as $mapping) {
                foreach ($this->variants->literalHints($mapping) as $old => $new) {
                    if (!str_contains($string->value, $old)) {
                        continue;
                    }

                    $warnings[] = new Warning(
                        message: $this->message($old, str_replace($old, $new, $string->value)),
                        file: $context->path,
                        line: $this->nodeLine($context, $string),
                    );
                }
            }
        }

        /** @var list<Name> $names */
        $names = $finder->findInstanceOf($context->ast, Name::class);
        foreach ($names as $name) {
            $reference = $name->getLast();
            foreach ($analysisContext->symbolMappings as $mapping) {
                $oldShortName = $this->variants->shortName($mapping->oldSymbol);
                if ($reference === $oldShortName || !str_contains($reference, $oldShortName)) {
                    continue;
                }

                if (RuleSupport::resolvedName($name) === $mapping->oldSymbol) {
                    continue;
                }

                $warnings[] = new Warning(
                    message: $this->message($reference, str_replace($oldShortName, $this->variants->shortName($mapping->newSymbol), $reference)),
                    file: $context->path,
                    line: $name->getStartLine(),
                );
            }
        }

        return $this->deduplicate($warnings);
    }

    private function message(string $old, string $new): string
    {
        return 'Semantic name "' . $old . '" resembles moved symbol; consider "' . $new . '". Not changed.';
    }

    private function nodeLine(PhpFileContext $context, Node $node): int
    {
        $offset = $node->getStartFilePos();
        if ($offset < 0) {
            return $node->getStartLine();
        }

        return $this->lineForOffset($context->content, $offset);
    }

    private function lineForOffset(string $content, int $offset): int
    {
        return mb_substr_count(mb_substr($content, 0, $offset), "\n") + 1;
    }

    /**
     * @param list<Warning> $warnings
     * @return list<Warning>
     */
    private function deduplicate(array $warnings): array
    {
        $seen = [];
        $deduplicated = [];

        foreach ($warnings as $warning) {
            $key = $warning->file . ':' . $warning->line . ':' . $warning->message;
            if (isset($seen[$key])) {
                continue;
            }

            $seen[$key] = true;
            $deduplicated[] = $warning;
        }

        return $deduplicated;
    }
}
