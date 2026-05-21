package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"refactorlah/internal/adapters"
)

func TestDryRunWritesNothing(t *testing.T) {
	root := copyFixture(t)
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:    "app/Services/Billing/InvoiceService.php",
		NewPath:    "app/Domain/Billing/InvoiceService.php",
		DryRun:     true,
		NoAdapters: true,
		Format:     FormatText,
	})
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	content, err := os.ReadFile(filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("namespace App\\Services\\Billing;")) {
		t.Fatal("fixture file changed during dry-run")
	}
}

func TestDefaultModeAppliesChanges(t *testing.T) {
	root := copyFixture(t)
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoAdapters:   true,
		NoValidation: true,
		Format:       FormatText,
	})
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "app", "Domain", "Billing", "InvoiceService.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func TestJSONOutputIsValidAndUnpolluted(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
		"--format=json",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !bytes.HasPrefix(bytes.TrimSpace(stdout.Bytes()), []byte("{")) {
		t.Fatalf("stdout is not JSON: %s", stdout.String())
	}
}

func TestApplyMovesFixtureFile(t *testing.T) {
	root := copyFixture(t)
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "app/Services/Billing/InvoiceService.php",
		NewPath:      "app/Domain/Billing/InvoiceService.php",
		Apply:        true,
		NoAdapters:   true,
		NoValidation: true,
		Format:       FormatText,
	})
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "app", "Domain", "Billing", "InvoiceService.php")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}
}

func TestNoAdaptersSkipsUnavailableAdapter(t *testing.T) {
	root := copyFixture(t)
	command := NewCommand()

	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:    "app/Services/Billing/InvoiceService.php",
		NewPath:    "app/Domain/Billing/InvoiceService.php",
		DryRun:     true,
		Apply:      false,
		NoAdapters: true,
		Format:     FormatText,
	})
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if len(report.AutoDetectedAdapters) != 0 {
		t.Fatalf("expected no adapters, got %v", report.AutoDetectedAdapters)
	}
}

func TestHelpShowsUsageWithoutError(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Commands:") {
		t.Fatalf("expected root command list, got: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "refactorlah <old-path> <new-path>") {
		t.Fatalf("did not expect shorthand usage in help: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got: %s", stderr.String())
	}
}

func TestNoArgsShowsUsageAndError(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), nil, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "error: expected command") {
		t.Fatalf("expected missing-args error, got: %s", stderr.String())
	}
	if strings.Index(stderr.String(), "error: expected command") > strings.Index(stderr.String(), "Usage:") {
		t.Fatalf("expected error before usage, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Commands:") {
		t.Fatalf("expected command list, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestInvalidFlagShowsErrorAboveUsage(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"move", "--apply"}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "error: flag provided but not defined: -apply") {
		t.Fatalf("expected invalid flag error, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stderr.String())
	}
	if strings.Index(stderr.String(), "error: flag provided but not defined: -apply") > strings.Index(stderr.String(), "Usage:") {
		t.Fatalf("expected error before usage, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestMoveHelpShowsMoveOptions(t *testing.T) {
	command := NewRootCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"move", "--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "--require-clean-worktree") {
		t.Fatalf("expected move options in help: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "--require-clean ") {
		t.Fatalf("did not expect old require-clean flag in help: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got: %s", stderr.String())
	}
}

func TestMoveSubcommandDelegatesToMoveCommand(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Mode: dry") {
		t.Fatalf("expected dry-run output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandSupportsUseListPairs(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-list",
		"app/Services/Billing/InvoiceService.php,app/Domain/Billing/InvoiceService.php",
		"tests/Feature/BillingTest.php,tests/Feature/BillingTestMoved.php",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected first move in output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "tests/Feature/BillingTest.php -> tests/Feature/BillingTestMoved.php") {
		t.Fatalf("expected second move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandAllowsLaterPairInsideEarlierTarget(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-list",
		"app/Services/Billing,app/Domain/Billing",
		"app/Domain/Billing/InvoiceService.php,app/Domain/Billing/BillingService.php",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/BillingService.php") {
		t.Fatalf("expected refined move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandSupportsUseFile(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "moves.txt"), []byte("app/Services/Billing/InvoiceService.php,app/Domain/Billing/InvoiceService.php\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"--use-file",
		"moves.txt",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected move in output, got: %s", stdout.String())
	}
}

func TestMoveSubcommandExpandsWildcardPairs(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"move",
		"--dry",
		"app/Services/Billing/*Service.php",
		"app/Domain/Billing/*Service.php",
		"--no-adapters",
	}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d stderr=%s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php") {
		t.Fatalf("expected wildcard-expanded move in output, got: %s", stdout.String())
	}
}

func TestDirectMoveWithoutCommandIsRejected(t *testing.T) {
	root := copyFixture(t)
	command := NewRootCommand()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{
		"app/Services/Billing/InvoiceService.php",
		"app/Domain/Billing/InvoiceService.php",
	}, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "error: unknown command \"app/Services/Billing/InvoiceService.php\"") {
		t.Fatalf("expected unknown-command error, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func TestReplacementTargetPathsRemapMovedFilesAndDeduplicate(t *testing.T) {
	paths := replacementTargetPaths(
		map[string]string{
			"app/Services/Billing/InvoiceService.php": "app/Domain/Billing/InvoiceService.php",
		},
		[]adapters.Replacement{
			{File: "app/Services/Billing/InvoiceService.php"},
			{File: "app/Services/Billing/InvoiceService.php"},
			{File: "app/Http/Controllers/InvoiceController.php"},
		},
	)

	if len(paths) != 2 {
		t.Fatalf("expected 2 staged paths, got %d: %#v", len(paths), paths)
	}
	if paths[0] != "app/Domain/Billing/InvoiceService.php" || paths[1] != "app/Http/Controllers/InvoiceController.php" {
		t.Fatalf("unexpected staged paths: %#v", paths)
	}
}

func TestApplyWithPHPAdapterKeepsImportsBeforeDeclarations(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\\\":"src/"}}}`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatch.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceLineCollection;

final readonly class InvoiceBatch
{
    public function __construct(
        public string $edition,
        public InvoiceFilter $range,
        public InvoiceTotals $stats,
        public InvoiceLineCollection $documents,
    ) {}
}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Archive", "Domain", "InvoiceLineCollection.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Archive\Domain;

final class InvoiceLineCollection {}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatchRepository.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Domain;

use App\Customer\Domain\CustomerId;

interface InvoiceBatchRepository
{
    public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
}
`)

	adapterPath, err := filepath.Abs(filepath.Join("..", "..", "adapters", "php", "bin", "refactorlah-php"))
	if err != nil {
		t.Fatal(err)
	}
	previousAdapterPath := os.Getenv("REFACTORLAH_PHP_ADAPTER")
	if err := os.Setenv("REFACTORLAH_PHP_ADAPTER", adapterPath); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if previousAdapterPath == "" {
			_ = os.Unsetenv("REFACTORLAH_PHP_ADAPTER")
			return
		}
		_ = os.Setenv("REFACTORLAH_PHP_ADAPTER", previousAdapterPath)
	}()

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Billing/Domain/InvoiceBatch.php",
		NewPath:      "src/Billing/Archive/Domain/InvoiceBatch.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	})
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	movedFile, err := os.ReadFile(filepath.Join(root, "src", "Billing", "Archive", "Domain", "InvoiceBatch.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(movedFile), "namespace App\\Billing\\Archive\\Domain;\n\nuse App\\Billing\\Domain\\InvoiceFilter;\nuse App\\Billing\\Domain\\InvoiceTotals;\n\nfinal readonly class InvoiceBatch") {
		t.Fatalf("expected imports before moved class declaration, got:\n%s", string(movedFile))
	}

	repositoryFile, err := os.ReadFile(filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatchRepository.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(repositoryFile), "use App\\Customer\\Domain\\CustomerId;\nuse App\\Billing\\Archive\\Domain\\InvoiceBatch;\n\ninterface InvoiceBatchRepository") {
		t.Fatalf("expected inserted import inside import block, got:\n%s", string(repositoryFile))
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sourceRoot := filepath.Join("..", "..", "adapters", "php", "tests", "fixtures", "php-basic")
	if runtime.GOOS == "windows" {
		sourceRoot = filepath.Clean(sourceRoot)
	}
	err := filepath.Walk(sourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
	if err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return root
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
