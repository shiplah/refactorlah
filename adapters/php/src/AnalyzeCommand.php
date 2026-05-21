<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter;

use PhpParser\Error;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use Refactorlah\PhpAdapter\Composer\ComposerConfigReader;
use Refactorlah\PhpAdapter\Files\FileCollector;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpCandidateFileSelector;
use Refactorlah\PhpAdapter\Php\PhpFileCollector;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\PhpReferenceScanner;
use Refactorlah\PhpAdapter\Php\PhpSymbolScanner;
use Refactorlah\PhpAdapter\Php\Psr4NamespaceResolver;
use Refactorlah\PhpAdapter\Php\SymbolMapping;
use Refactorlah\PhpAdapter\Project\ProjectContextResolver;
use Refactorlah\PhpAdapter\Protocol\Request;
use Refactorlah\PhpAdapter\Protocol\Response;
use Refactorlah\PhpAdapter\Twig\TwigConfigReader;
use Refactorlah\PhpAdapter\Twig\TwigReferenceScanner;
use Refactorlah\PhpAdapter\Twig\TwigTemplateMapper;

use function array_filter;
use function array_map;
use function array_merge;
use function array_values;
use function file_get_contents;
use function fwrite;
use function getcwd;
use function json_decode;
use function json_encode;
use function str_ends_with;
use function stream_get_contents;

/**
 * @phpstan-import-type SymbolMappingArray from \Refactorlah\PhpAdapter\Php\SymbolMapping
 */
final class AnalyzeCommand
{
    /**
     * @param list<string> $argv
     */
    public function run(array $argv): int
    {
        if (($argv[1] ?? '') !== 'analyze') {
            fwrite(STDERR, "usage: refactorlah-php analyze\n");
            return 2;
        }

        try {
            $request = Request::fromArray($this->decodeRequestPayload((string) stream_get_contents(STDIN)));
            $projectRoot = getcwd() ?: '.';
            $projectContext = (new ProjectContextResolver())->resolve($projectRoot, $request->moves);
            $subRootMoves = array_map(
                static fn(array $move): array => [
                    'oldPath' => $projectContext->toSubRootRelative($move['oldPath']),
                    'newPath' => $projectContext->toSubRootRelative($move['newPath']),
                    'tracked' => $move['tracked'],
                ],
                $request->moves,
            );

            $composerReader = new ComposerConfigReader();
            $psr4Map = $composerReader->readPsr4Map($projectContext->absoluteRoot);

            $symbolScanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
            [$symbolMappings, $warnings] = $symbolScanner->scan($projectContext->absoluteRoot, $psr4Map, $subRootMoves);
            $analysisMappings = $symbolMappings;

            foreach ($symbolMappings as $index => $mapping) {
                $symbolMappings[$index] = new SymbolMapping(
                    kind: $mapping->kind,
                    oldPath: $projectContext->toProjectRelative($mapping->oldPath),
                    newPath: $projectContext->toProjectRelative($mapping->newPath),
                    oldSymbol: $mapping->oldSymbol,
                    newSymbol: $mapping->newSymbol,
                    oldNamespace: $mapping->oldNamespace,
                    newNamespace: $mapping->newNamespace,
                    shortName: $mapping->shortName,
                );
            }

            foreach ($warnings as $index => $warning) {
                $warnings[$index] = new \Refactorlah\PhpAdapter\Warning\Warning(
                    message: $warning->message,
                    file: '' !== $warning->file ? $projectContext->toProjectRelative($warning->file) : '',
                    line: $warning->line,
                );
            }

            $pathMappings = $request->includeTwig
                ? (new TwigTemplateMapper())->deriveMappings(
                    $subRootMoves,
                    (new TwigConfigReader())->read($projectContext->absoluteRoot)
                )
                : [];

            foreach ($pathMappings as $index => $mapping) {
                $pathMappings[$index]['oldPath'] = $projectContext->toProjectRelative($mapping['oldPath']);
                $pathMappings[$index]['newPath'] = $projectContext->toProjectRelative($mapping['newPath']);
            }

            $symbolMappingIndex = [];
            foreach ($analysisMappings as $mapping) {
                $symbolMappingIndex[$mapping->oldSymbol] = $mapping;
            }
            $analysisContext = new AnalysisContext(
                symbolMappings: $symbolMappingIndex
            );

            $replacements = [];

            if ($request->includePhp) {
                $phpFiles = (new PhpFileCollector(new FileCollector()))->collect($projectContext->absoluteRoot);
                $candidateFiles = (new PhpCandidateFileSelector())->select(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $phpFiles,
                    symbolMappings: $analysisMappings,
                    movedPhpFiles: array_map(
                        static fn(array $move): string => $move['oldPath'],
                        array_values(array_filter(
                            $subRootMoves,
                            static fn(array $move): bool => str_ends_with($move['oldPath'], '.php'),
                        )),
                    ),
                );
                if ([] !== $candidateFiles) {
                    $phpContexts = $this->parsePhpFiles($projectContext->absoluteRoot, $candidateFiles);
                    $scanner = new PhpReferenceScanner();
                    [$phpReplacements, $phpWarnings] = $scanner->scan($phpContexts, $analysisContext);
                    foreach ($phpReplacements as $index => $replacement) {
                        $phpReplacements[$index] = new \Refactorlah\PhpAdapter\Replacement\Replacement(
                            file: $projectContext->toProjectRelative($replacement->file),
                            start: $replacement->start,
                            end: $replacement->end,
                            replacement: $replacement->replacement,
                            reason: $replacement->reason,
                            rule: $replacement->rule,
                        );
                    }
                    foreach ($phpWarnings as $index => $warning) {
                        $phpWarnings[$index] = new \Refactorlah\PhpAdapter\Warning\Warning(
                            message: $warning->message,
                            file: '' !== $warning->file ? $projectContext->toProjectRelative($warning->file) : '',
                            line: $warning->line,
                        );
                    }
                    $replacements = array_merge($replacements, $phpReplacements);
                    $warnings = array_merge($warnings, $phpWarnings);
                }
            }

            if ($request->includeTwig) {
                $twigScanner = new TwigReferenceScanner(new FileCollector());
                $registry = new \Refactorlah\PhpAdapter\Twig\TwigRuleRegistry();
                [$twigReplacements, $twigWarnings] = $registry->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $twigScanner->collectConfigFiles($projectContext->absoluteRoot),
                    twigFiles: $twigScanner->collectTwigFiles($projectContext->absoluteRoot),
                    pathMappings: $pathMappings,
                );
                foreach ($twigReplacements as $index => $replacement) {
                    $twigReplacements[$index] = new \Refactorlah\PhpAdapter\Replacement\Replacement(
                        file: $projectContext->toProjectRelative($replacement->file),
                        start: $replacement->start,
                        end: $replacement->end,
                        replacement: $replacement->replacement,
                        reason: $replacement->reason,
                        rule: $replacement->rule,
                    );
                }
                foreach ($twigWarnings as $index => $warning) {
                    $twigWarnings[$index] = new \Refactorlah\PhpAdapter\Warning\Warning(
                        message: $warning->message,
                        file: '' !== $warning->file ? $projectContext->toProjectRelative($warning->file) : '',
                        line: $warning->line,
                    );
                }
                $replacements = array_merge($replacements, $twigReplacements);
                $warnings = array_merge($warnings, $twigWarnings);
            }

            echo json_encode(new Response(
                symbolMappings: $this->serializeSymbolMappings($symbolMappings),
                pathMappings: $pathMappings,
                replacements: $replacements,
                warnings: $warnings,
                errors: [],
            ), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);

            return 0;
        } catch (\Throwable $throwable) {
            fwrite(STDERR, $throwable->getMessage() . PHP_EOL);
            echo json_encode(new Response([], [], [], [], [$throwable->getMessage()]), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
            return 1;
        }
    }

    /**
     * @return array<string,mixed>
     */
    private function decodeRequestPayload(string $payload): array
    {
        $decoded = json_decode($payload, true, flags: JSON_THROW_ON_ERROR);
        if (!\is_array($decoded)) {
            throw new \RuntimeException('adapter request must decode to an object');
        }

        $normalized = [];
        foreach ($decoded as $key => $value) {
            if (!\is_string($key)) {
                continue;
            }

            $normalized[$key] = $value;
        }

        return $normalized;
    }

    /**
     * @param list<SymbolMapping> $symbolMappings
     * @return list<SymbolMappingArray>
     */
    private function serializeSymbolMappings(array $symbolMappings): array
    {
        return array_map(static fn(SymbolMapping $mapping): array => $mapping->toArray(), $symbolMappings);
    }

    /**
     * @param list<string> $files
     * @return list<PhpFileContext>
     */
    private function parsePhpFiles(string $projectRoot, array $files): array
    {
        $parser = (new ParserFactory())->createForNewestSupportedVersion();
        $contexts = [];

        foreach ($files as $file) {
            $content = (string) file_get_contents($projectRoot . '/' . $file);
            try {
                $ast = array_values($parser->parse($content) ?? []);
                $traverser = new NodeTraverser();
                $traverser->addVisitor(new NameResolver(options: ['preserveOriginalNames' => true]));
                $resolved = array_values($traverser->traverse($ast));
                \Refactorlah\PhpAdapter\Php\RuleSupport::attachParents($resolved);
                $contexts[] = new PhpFileContext($file, $content, $resolved);
            } catch (Error) {
                continue;
            }
        }

        return $contexts;
    }
}
