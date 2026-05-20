<?php

declare(strict_types=1);

namespace Refactorlah\PhpAdapter\Twig;

final class TwigConfigReader
{
    public function read(string $projectRoot): TwigPathConfiguration
    {
        $roots = [];
        $configPath = $projectRoot . '/config/packages/twig.yaml';

        if (!is_file($configPath)) {
            if (is_dir($projectRoot . '/templates')) {
                $roots[] = new TwigPathRoot('templates');
            }

            return new TwigPathConfiguration($roots);
        }

        $lines = file($configPath, FILE_IGNORE_NEW_LINES) ?: [];
        $inTwigBlock = false;
        $inPathsBlock = false;

        foreach ($lines as $line) {
            if (trim($line) === 'twig:') {
                $inTwigBlock = true;
                $inPathsBlock = false;
                continue;
            }

            if ($inTwigBlock && preg_match('/^[^\s]/', $line) === 1) {
                $inTwigBlock = false;
                $inPathsBlock = false;
            }

            if (!$inTwigBlock) {
                continue;
            }

            if (preg_match('/^\s{2}default_path:\s*[\'"]?%kernel\.project_dir%\/([^\'"]+)[\'"]?\s*$/', $line, $matches) === 1) {
                $roots[] = new TwigPathRoot(trim($matches[1], '/'));
                continue;
            }

            if (preg_match('/^\s{2}paths:\s*$/', $line) === 1) {
                $inPathsBlock = true;
                continue;
            }

            if ($inPathsBlock && preg_match('/^\s{4}[\'"]?%kernel\.project_dir%\/([^\'"]+)[\'"]?\s*:\s*([A-Za-z0-9_]+)\s*$/', $line, $matches) === 1) {
                $roots[] = new TwigPathRoot(trim($matches[1], '/'), $matches[2]);
                continue;
            }

            if ($inPathsBlock && preg_match('/^\s{2}\S/', $line) === 1) {
                $inPathsBlock = false;
            }
        }

        return new TwigPathConfiguration($roots);
    }
}
