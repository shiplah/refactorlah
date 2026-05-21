package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
	command := NewCommand()

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
		"--dry-run",
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
	command := NewCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), []string{"--help"}, &stdout, &stderr)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "--apply") {
		t.Fatalf("did not expect removed --apply flag in help: %s", stdout.String())
	}
	if strings.Contains(stdout.String(), "--allow-dirty") || strings.Contains(stdout.String(), "--allow-no-git") {
		t.Fatalf("did not expect old allow-* flags in help: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "--require-clean") || !strings.Contains(stdout.String(), "--require-git") {
		t.Fatalf("expected new require-* flags in help: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got: %s", stderr.String())
	}
}

func TestNoArgsShowsUsageAndError(t *testing.T) {
	command := NewCommand()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := command.Run(t.Context(), nil, &stdout, &stderr)
	if exitCode != ExitInvalidArguments {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("expected usage output, got: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "expected <old-path> and <new-path>") {
		t.Fatalf("expected missing-args error, got: %s", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected empty stdout, got: %s", stdout.String())
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	sourceRoot := filepath.Join("..", "..", "tests", "fixtures", "php-basic")
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
