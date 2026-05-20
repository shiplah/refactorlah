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
use Refactorlah\PhpAdapter\Php\PhpFileCollector;
use Refactorlah\PhpAdapter\Php\PhpFileContext;
use Refactorlah\PhpAdapter\Php\PhpReferenceScanner;
use Refactorlah\PhpAdapter\Php\PhpSymbolScanner;
use Refactorlah\PhpAdapter\Php\Psr4NamespaceResolver;
use Refactorlah\PhpAdapter\Protocol\Request;
use Refactorlah\PhpAdapter\Protocol\Response;
use Refactorlah\PhpAdapter\Php\SymbolMapping;
use Refactorlah\PhpAdapter\Twig\TwigReferenceScanner;
use Refactorlah\PhpAdapter\Twig\TwigTemplateMapper;
use Refactorlah\PhpAdapter\Twig\TwigWorkerRegistry;

final class AnalyzeCommand
{
    public function run(array $argv): int
    {
        if (($argv[1] ?? '') !== 'analyze') {
            fwrite(STDERR, "usage: refactorlah-php analyze\n");
            return 2;
        }

        try {
            $request = Request::fromArray(json_decode((string) stream_get_contents(STDIN), true, flags: JSON_THROW_ON_ERROR));
            $projectRoot = getcwd() ?: '.';

            $composerReader = new ComposerConfigReader();
            $psr4Map = $composerReader->readPsr4Map($projectRoot);

            $symbolScanner = new PhpSymbolScanner(new Psr4NamespaceResolver());
            [$symbolMappings, $warnings] = $symbolScanner->scan($projectRoot, $psr4Map, $request->moves);

            $pathMappings = $request->includeTwig
                ? (new TwigTemplateMapper())->deriveMappings($request->moves)
                : [];

            $symbolMappingIndex = [];
            foreach ($symbolMappings as $mapping) {
                $symbolMappingIndex[$mapping->oldSymbol] = $mapping;
            }
            $analysisContext = new AnalysisContext(
                symbolMappings: $symbolMappingIndex
            );

            $replacements = [];

            if ($request->includePhp) {
                $phpFiles = (new PhpFileCollector(new FileCollector()))->collect($projectRoot);
                $phpContexts = $this->parsePhpFiles($projectRoot, $phpFiles);
                $scanner = new PhpReferenceScanner();
                [$phpReplacements, $phpWarnings] = $scanner->scan($phpContexts, $analysisContext);
                $replacements = array_merge($replacements, $phpReplacements);
                $warnings = array_merge($warnings, $phpWarnings);
            }

            if ($request->includeTwig) {
                $twigScanner = new TwigReferenceScanner(new FileCollector());
                $registry = new TwigWorkerRegistry();
                [$twigReplacements, $twigWarnings] = $registry->scan(
                    projectRoot: $projectRoot,
                    files: $twigScanner->collectConfigFiles($projectRoot),
                    twigFiles: $twigScanner->collectTwigFiles($projectRoot),
                    pathMappings: $pathMappings,
                );
                $replacements = array_merge($replacements, $twigReplacements);
                $warnings = array_merge($warnings, $twigWarnings);
            }

            echo json_encode(new Response(
                symbolMappings: array_map(static fn ($mapping) => $mapping->toArray(), $symbolMappings),
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
                $ast = $parser->parse($content) ?? [];
                $traverser = new NodeTraverser();
                $traverser->addVisitor(new NameResolver(options: ['preserveOriginalNames' => true]));
                $resolved = $traverser->traverse($ast);
                \Refactorlah\PhpAdapter\Php\WorkerSupport::attachParents($resolved);
                $contexts[] = new PhpFileContext($file, $content, $resolved);
            } catch (Error) {
                continue;
            }
        }

        return $contexts;
    }
}
