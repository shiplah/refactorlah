package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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
		AllowDirty:   true,
		AllowNoGit:   true,
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
