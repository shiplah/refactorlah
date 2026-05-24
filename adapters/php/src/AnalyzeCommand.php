<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter;

use Refactorlah\PhpAdapter\Composer\ComposerConfigReader;
use Refactorlah\PhpAdapter\Config\PathMappingCollection;
use Refactorlah\PhpAdapter\Config\PathMappingFactory;
use Refactorlah\PhpAdapter\Config\StaticImportReferenceScanner;
use Refactorlah\PhpAdapter\Files\FileCollector;
use Refactorlah\PhpAdapter\Php\AnalysisContext;
use Refactorlah\PhpAdapter\Php\MovedPhpFileSelector;
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

use function array_map;
use function array_merge;
use function fwrite;
use function getcwd;
use function is_array;
use function is_string;
use function json_decode;
use function json_encode;
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
                include: array_map($projectContext->toSubRootRelative(...), $request->options->scanInclude),
                exclude: array_map($projectContext->toSubRootRelative(...), $request->options->scanExclude),
            );
            $subRootMoves = $request->moves->toSubRootRelative($projectContext);

            $composerReader = new ComposerConfigReader();
            $psr4Map = $composerReader->readPsr4Map($projectContext->absoluteRoot);

            $symbolScanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
            $symbolScan = $symbolScanner->scan($projectContext->absoluteRoot, $psr4Map, $subRootMoves);
            $symbolMappings = $symbolScan->symbolMappings;
            $warnings = $symbolScan->warnings;
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

            $pathMappings = $request->options->includeTwig
                ? (new TwigTemplateMapper())->deriveMappings(
                    $subRootMoves,
                    (new TwigConfigReader())->read($projectContext->absoluteRoot)
                )
                : PathMappingCollection::empty();
            $configPathMappings = (new PathMappingFactory())->fromMove(
                $projectContext->toSubRootRelative($request->oldPath),
                $projectContext->toSubRootRelative($request->newPath),
            );

            $pathMappings = $pathMappings->toProjectRelative($projectContext);
            $projectPathMappings = $configPathMappings->toProjectRelative($projectContext);

            $symbolMappingIndex = [];
            foreach ($analysisMappings as $mapping) {
                $symbolMappingIndex[$mapping->oldSymbol] = $mapping;
            }
            $analysisContext = new AnalysisContext(
                symbolMappings: $symbolMappingIndex
            );

            $replacements = [];

            if ($request->options->includePhp) {
                $phpFiles = $scanPolicy->filter((new PhpFileCollector(new FileCollector()))->collect($projectContext->absoluteRoot));
                $candidateFiles = (new PhpCandidateFileSelector())->select(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $phpFiles,
                    symbolMappings: $analysisMappings,
                    movedPhpFiles: (new MovedPhpFileSelector())->oldPaths($subRootMoves),
                );
                if ([] !== $candidateFiles) {
                    $phpContexts = (new PhpFileParser())->parse($projectContext->absoluteRoot, $candidateFiles);
                    $scanner = new PhpReferenceScanner();
                    $phpScan = $scanner->scan($phpContexts, $analysisContext);
                    $phpReplacements = $phpScan->replacements;
                    $phpWarnings = $phpScan->warnings;
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

            if ($request->options->includeTwig) {
                $twigScanner = new TwigReferenceScanner(new FileCollector());
                $registry = new \Refactorlah\PhpAdapter\Symfony\Twig\TwigRuleRegistry();
                $twigScan = $registry->scan(
                    projectRoot: $projectContext->absoluteRoot,
                    files: $scanPolicy->filter($twigScanner->collectConfigFiles($projectContext->absoluteRoot)),
                    twigFiles: $scanPolicy->filter($twigScanner->collectTwigFiles($projectContext->absoluteRoot)),
                    pathMappings: $pathMappings,
                );
                $twigReplacements = $twigScan->replacements;
                $twigWarnings = $twigScan->warnings;
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
            $pathMappings = $pathMappings->merge($projectPathMappings);

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
            echo json_encode(new Response([], PathMappingCollection::empty(), [], [], [$throwable->getMessage()]), JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES);
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
