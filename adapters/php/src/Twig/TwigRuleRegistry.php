<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

use Refactorlah\PhpAdapter\Replacement\Replacement;
use Refactorlah\PhpAdapter\Warning\Warning;

use function array_merge;
use function array_unique;
use function array_values;
use function basename;
use function file_get_contents;
use function mb_substr;
use function mb_substr_count;
use function mb_trim;
use function preg_match_all;
use function str_contains;
use function str_ends_with;

final class TwigRuleRegistry
{
    /**
     * @param list<string> $files
     * @param list<string> $twigFiles
     * @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings
     * @return array{0:list<Replacement>,1:list<Warning>}
     */
    public function scan(string $projectRoot, array $files, array $twigFiles, array $pathMappings): array
    {
        if ([] === $pathMappings) {
            return [[], []];
        }

        /** @var list<\Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule> $twigRules */
        $twigRules = [
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigIncludeReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigExtendsReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigEmbedReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigUseReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigImportReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigFromReplacementRule(),
        ];
        /** @var list<\Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule> $phpRules */
        $phpRules = [
            new \Refactorlah\PhpAdapter\Twig\Rules\SymfonyRenderTemplateReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\SymfonyTemplateAttributeReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\TwigComponentTemplateAttributeReplacementRule(),
        ];
        /** @var list<\Refactorlah\PhpAdapter\Twig\Rules\AbstractTwigStringReplacementRule> $yamlRules */
        $yamlRules = [
            new \Refactorlah\PhpAdapter\Twig\Rules\YamlTwigTemplateReplacementRule(),
            new \Refactorlah\PhpAdapter\Twig\Rules\YamlTwigComponentTemplateDirectoryReplacementRule(),
        ];

        $replacements = [];
        $warnings = [];

        foreach ($twigFiles as $file) {
            $content = (string) file_get_contents($projectRoot . '/' . $file);
            if (!$this->containsMappedReference($content, $pathMappings)) {
                continue;
            }
            foreach ($pathMappings as $mapping) {
                foreach ($twigRules as $rule) {
                    $replacements = array_merge($replacements, $rule->collect($file, $content, $mapping));
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
                $rules = str_ends_with($file, '.php') ? $phpRules : $yamlRules;
                foreach ($rules as $rule) {
                    $replacements = array_merge($replacements, $rule->collect($file, $content, $mapping));
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
                    line: mb_substr_count(mb_substr($content, 0, $offset), "\n") + 1,
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
            if ('' === mb_trim($value)) {
                continue;
            }
            if (!$this->containsIndicator($value, $indicators)) {
                continue;
            }
            $warnings[] = new Warning(
                message: 'Dynamic Twig template path detected; not changed.',
                file: $file,
                line: mb_substr_count(mb_substr($content, 0, $offset), "\n") + 1,
            );
        }

        return $warnings;
    }

    /** @param list<array{kind:string,oldPath:string,newPath:string,oldReference:string,newReference:string}> $pathMappings */
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

    /** @param list<string> $indicators */
    private function containsIndicator(string $value, array $indicators): bool
    {
        foreach ($indicators as $indicator) {
            if ('' !== $indicator && str_contains($value, $indicator)) {
                return true;
            }
        }

        return false;
    }
}
