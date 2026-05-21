<?php

declare(strict_types=1);

require __DIR__ . '/bootstrap.php';
require __DIR__ . '/Unit/PhpCandidateFileSelectorTest.php';
require __DIR__ . '/Unit/PhpSymbolScannerTest.php';
require __DIR__ . '/Unit/PhpRulesTest.php';
require __DIR__ . '/Unit/TwigRulesTest.php';
require __DIR__ . '/Unit/AnalyzeCommandTest.php';

exit(run_all_tests());
