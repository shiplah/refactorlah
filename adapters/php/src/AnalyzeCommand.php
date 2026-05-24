<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter;

use Refactorlah\PhpAdapter\Composer\ComposerConfigReader;
use Refactorlah\PhpAdapter\Config\PathMappingFactory;
use Refactorlah\PhpAdapter\Config\StaticImportReferenceScanner;
use Refactorlah\PhpAdapter\Files\FileCollector;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\PhpCandidateFileSelector;
use Refactorlah\PhpAdapter\Php\PhpFileCollector;
use Refactorlah\PhpAdapter\Php\PhpFileParser;
use Refactorlah\PhpAdapter\Php\PhpReferenceScanner;
use Refactorlah\PhpAdapter\Php\PhpSymbolScanner;
use Refactorlah\PhpAdapter\Php\Psr4NamespaceResolver;
use Refactorlah\PhpAdapter\Php\SemanticRenameHintScanner;
use Refactorlah\PhpAdapter\Php\SymbolMapping;
use Refactorlah\PhpAdapter\Project\ProjectContextResolver;
use Refactorlah\PhpAdapter\Project\ScanPolicy;
use Refactorlah\PhpAdapter\Protocol\Request;
use Refactorlah\PhpAdapter\Protocol\Response;
use Refactorlah\PhpAdapter\Symfony\Core\YamlAssetMapperPathReferenceScanner;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigConfigReader;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigReferenceScanner;
use Refactorlah\PhpAdapter\Symfony\Twig\TwigTemplateMapper;
use Refactorlah\PhpAdapter\Symfony\Twig\YamlComponentNamespaceReferenceScanner;

use function array_filter;
use function array_map;
use function array_merge;
use function array_values;
use function fwrite;
use function getcwd;
use function is_array;
use function is_string;
use function json_decode;
use function json_encode;
use function str_ends_with;
use function stream_get_contents;

/**
 * @phpstan-import-type SymbolMappingArray from \Refactorlah\PhpAdapter\Php\SymbolMapping
 */
final class AnalyzeCommand
{
    /** @param list<string> $argv */
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
            $scanPolicy = new ScanPolicy(
                include: array_map($projectContext->toSubRootRelative(...), $request->scanInclude),
                exclude: array_map($projectContext->toSubRootRelative(...), $request->scanExclude),
            );
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
            $configPathMappings = (new PathMappingFactory())->fromMove(
                $projectContext->toSubRootRelative($request->oldPath),
                $projectContext->toSubRootRelative($request->newPath),
            );

            foreach ($pathMappings as $index => $mapping) {
                $pathMappings[$index]['oldPath'] = $projectContext->toProjectRelative($mapping['oldPath']);
                $pathMappings[$index]['newPath'] = $projectContext->toProjectRelative($mapping['newPath']);
            }
            $projectPathMappings = $configPathMappings;
            foreach ($projectPathMappings as $index => $mapping) {
                $projectPathMappings[$index]['oldPath'] = $projectContext->toProjectRelative($mapping['oldPath']);
                $projectPathMappings[$index]['newPath'] = $projectContext->toProjectRelative($mapping['newPath']);
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
                $phpFiles = $scanPolicy->filter((new PhpFileCollector(new FileCollector()))->collect($projectContext->absoluteRoot));
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
                    $phpContexts = (new PhpFileParser())->parse($projectContext->absoluteRoot, $candidateFiles);
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

                $yamlReplacements = (new YamlComponentNamespaceReferenceScanner())->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter(
                        (new FileCollector())->collect($projectContext->absoluteRoot, ['yaml', 'yml']),
                    ),
                    symbolMappings: $analysisMappings,
                );
                foreach ($yamlReplacements as $index => $replacement) {
                    $yamlReplacements[$index] = new \Refactorlah\PhpAdapter\Replacement\Replacement(
                        file: $projectContext->toProjectRelative($replacement->file),
                        start: $replacement->start,
                        end: $replacement->end,
                        replacement: $replacement->replacement,
                        reason: $replacement->reason,
                        rule: $replacement->rule,
                    );
                }
                $replacements = array_merge($replacements, $yamlReplacements);

                $pathReplacements = (new YamlAssetMapperPathReferenceScanner())->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter(
                        (new FileCollector())->collect($projectContext->absoluteRoot, ['yaml', 'yml']),
                    ),
                    pathMappings: $configPathMappings,
                );
                foreach ($pathReplacements as $index => $replacement) {
                    $pathReplacements[$index] = new \Refactorlah\PhpAdapter\Replacement\Replacement(
                        file: $projectContext->toProjectRelative($replacement->file),
                        start: $replacement->start,
                        end: $replacement->end,
                        replacement: $replacement->replacement,
                        reason: $replacement->reason,
                        rule: $replacement->rule,
                    );
                }
                $replacements = array_merge($replacements, $pathReplacements);

                $staticImportReplacements = (new StaticImportReferenceScanner())->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter(
                        (new FileCollector())->collect($projectContext->absoluteRoot, ['js', 'jsx', 'ts', 'tsx', 'mjs', 'cjs', 'css']),
                    ),
                    moves: $subRootMoves,
                );
                foreach ($staticImportReplacements as $index => $replacement) {
                    $staticImportReplacements[$index] = new \Refactorlah\PhpAdapter\Replacement\Replacement(
                        file: $projectContext->toProjectRelative($replacement->file),
                        start: $replacement->start,
                        end: $replacement->end,
                        replacement: $replacement->replacement,
                        reason: $replacement->reason,
                        rule: $replacement->rule,
                    );
                }
                $replacements = array_merge($replacements, $staticImportReplacements);

                $semanticHintWarnings = (new SemanticRenameHintScanner())->scanTextFiles(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter(
                        (new FileCollector())->collect($projectContext->absoluteRoot, ['yaml', 'yml', 'xml', 'neon']),
                    ),
                    symbolMappings: $analysisMappings,
                );
                foreach ($semanticHintWarnings as $index => $warning) {
                    $semanticHintWarnings[$index] = new \Refactorlah\PhpAdapter\Warning\Warning(
                        message: $warning->message,
                        file: '' !== $warning->file ? $projectContext->toProjectRelative($warning->file) : '',
                        line: $warning->line,
                    );
                }
                $warnings = array_merge($warnings, $semanticHintWarnings);
            }

            if ($request->includeTwig) {
                $twigScanner = new TwigReferenceScanner(new FileCollector());
                $registry = new \Refactorlah\PhpAdapter\Symfony\Twig\TwigRuleRegistry();
                [$twigReplacements, $twigWarnings] = $registry->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter($twigScanner->collectConfigFiles($projectContext->absoluteRoot)),
                    twigFiles: $scanPolicy->filter($twigScanner->collectTwigFiles($projectContext->absoluteRoot)),
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
            $pathMappings = array_merge($pathMappings, $projectPathMappings);

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

    /** @return array<string,mixed> */
    private function decodeRequestPayload(string $payload): array
    {
        $decoded = json_decode($payload, true, flags: JSON_THROW_ON_ERROR);
        if (!is_array($decoded)) {
            throw new \RuntimeException('adapter request must decode to an object');
        }

        $normalized = [];
        foreach ($decoded as $key => $value) {
            if (!is_string($key)) {
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

}
