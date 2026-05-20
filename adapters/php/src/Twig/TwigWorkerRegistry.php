<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Twig\Workers\SymfonyRenderTemplateReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\SymfonyTemplateAttributeReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigEmbedReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigExtendsReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigFromReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigImportReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigIncludeReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\TwigUseReplacementWorker;
use Refactorlah\PhpAdapter\Twig\Workers\YamlTwigTemplateReplacementWorker;
use Refactorlah\PhpAdapter\Warning\Warning;

final class TwigWorkerRegistry
{
    /**
     * @param list<string> $files
     * @param list<string> $twigFiles
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     * @return array{0:list<Replacement>,1:list<Warning>}
     */
    public function scan(string $projectRoot, array $files, array $twigFiles, array $pathMappings): array
    {
        if ($pathMappings === []) {
            return [[], []];
        }

        $twigWorkers = [
            new TwigIncludeReplacementWorker(),
            new TwigExtendsReplacementWorker(),
            new TwigEmbedReplacementWorker(),
            new TwigUseReplacementWorker(),
            new TwigImportReplacementWorker(),
            new TwigFromReplacementWorker(),
        ];
        $phpWorkers = [
            new SymfonyRenderTemplateReplacementWorker(),
            new SymfonyTemplateAttributeReplacementWorker(),
        ];
        $yamlWorkers = [
            new YamlTwigTemplateReplacementWorker(),
        ];

        $replacements = [];
        $warnings = [];

        foreach ($twigFiles as $file) {
            $content = (string) file_get_contents($projectRoot . '/' . $file);
            if (!$this->containsMappedReference($content, $pathMappings)) {
                continue;
            }
            foreach ($pathMappings as $mapping) {
                foreach ($twigWorkers as $worker) {
                    $replacements = array_merge($replacements, $worker->collect($file, $content, $mapping));
                }
            }
            $warnings = array_merge($warnings, $this->twigWarnings($file, $content, $pathMappings));
        }

        foreach ($files as $file) {
            $content = (string) file_get_contents($projectRoot . '/' . $file);
            if (!$this->containsMappedReference($content, $pathMappings)) {
                continue;
            }
            foreach ($pathMappings as $mapping) {
                $workers = str_ends_with($file, '.php') ? $phpWorkers : $yamlWorkers;
                foreach ($workers as $worker) {
                    $replacements = array_merge($replacements, $worker->collect($file, $content, $mapping));
                }
            }
            if (str_ends_with($file, '.php')) {
                $warnings = array_merge($warnings, $this->phpWarnings($file, $content, $pathMappings));
            }
        }

        return [$replacements, $warnings];
    }

    /**
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     * @return list<Warning>
     */
    private function twigWarnings(string $file, string $content, array $pathMappings): array
    {
        $warnings = [];
        $indicators = $this->warningIndicators($pathMappings);
        foreach ([
            '/{%\s*include\s+([A-Za-z_][^%\s]*)/',
            '/{{\s*include\(\s*([A-Za-z_][^)]+)\)/',
            '/{%\s*extends\s+([A-Za-z_][^%\s]*)/',
        ] as $pattern) {
            if (!preg_match_all($pattern, $content, $matches, PREG_OFFSET_CAPTURE)) {
                continue;
            }
            foreach ($matches[1] as [$value, $offset]) {
                if (!$this->containsIndicator($value, $indicators)) {
                    continue;
                }
                $warnings[] = new Warning(
                    message: 'Dynamic Twig template path detected; not changed.',
                    file: $file,
                    line: substr_count(substr($content, 0, $offset), "\n") + 1,
                );
            }
        }

        return $warnings;
    }

    /**
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     * @return list<Warning>
     */
    private function phpWarnings(string $file, string $content, array $pathMappings): array
    {
        $warnings = [];
        $indicators = $this->warningIndicators($pathMappings);
        if (!preg_match_all('/->render(?:View)?\(\s*([^\'"][^,\)]*)/m', $content, $matches, PREG_OFFSET_CAPTURE)) {
            return [];
        }

        foreach ($matches[1] as [$value, $offset]) {
            if (trim($value) === '') {
                continue;
            }
            if (!$this->containsIndicator($value, $indicators)) {
                continue;
            }
            $warnings[] = new Warning(
                message: 'Dynamic Twig template path detected; not changed.',
                file: $file,
                line: substr_count(substr($content, 0, $offset), "\n") + 1,
            );
        }

        return $warnings;
    }

    /**
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     */
    private function containsMappedReference(string $content, array $pathMappings): bool
    {
        foreach ($pathMappings as $mapping) {
            if (str_contains($content, $mapping['oldReference'])) {
                return true;
            }
        }

        return false;
    }

    /**
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     * @return list<string>
     */
    private function warningIndicators(array $pathMappings): array
    {
        $indicators = [];
        foreach ($pathMappings as $mapping) {
            $indicators[] = $mapping['oldReference'];
            $indicators[] = basename($mapping['oldReference']);
        }

        return array_values(array_unique($indicators));
    }

    /**
     * @param list<string> $indicators
     */
    private function containsIndicator(string $value, array $indicators): bool
    {
        foreach ($indicators as $indicator) {
            if ($indicator !== '' && str_contains($value, $indicator)) {
                return true;
            }
        }

        return false;
    }
}
