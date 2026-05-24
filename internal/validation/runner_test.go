package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerReportsNewValidationFailures(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "vendor", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	phpstanPath := filepath.Join(binDir, "phpstan")
	script := "#!/bin/sh\nprintf 'existing failure\\n' >&2\nexit 1\n"
	if err := os.WriteFile(phpstanPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := NewRunner()
	results, err := runner.Run(t.Context(), root, RunOptions{})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected validation failure, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 validation result, got %d", len(results))
	}
	if results[0].Name != "phpstan" {
		t.Fatalf("expected phpstan result, got %#v", results[0])
	}
	if results[0].Message != newValidationFailureMessage {
		t.Fatalf("unexpected validation message: %#v", results[0])
	}
	if !strings.Contains(results[0].Stderr, "existing failure") {
		t.Fatalf("expected captured validator stderr, got %#v", results[0])
	}
}

func TestRunnerIgnoresUnchangedPreExistingValidationFailures(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "vendor", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	phpstanPath := filepath.Join(binDir, "phpstan")
	script := "#!/bin/sh\nprintf 'pre-existing architecture failure\\n' >&2\nexit 1\n"
	if err := os.WriteFile(phpstanPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := NewRunner()
	baseline := runner.Baseline(t.Context(), root, RunOptions{})
	results, err := runner.RunCompared(t.Context(), root, RunOptions{}, baseline)
	if err != nil {
		t.Fatalf("expected unchanged pre-existing failure not to fail refactor validation, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 validation result, got %d", len(results))
	}
	if results[0].Status != "unchanged-failure" {
		t.Fatalf("expected unchanged failure status, got %#v", results[0])
	}
	if results[0].Message != unchangedValidationFailureMessage {
		t.Fatalf("unexpected validation message: %#v", results[0])
	}
}

func TestRunnerReportsChangedValidationFailures(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "vendor", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	statePath := filepath.Join(root, "state")
	phpstanPath := filepath.Join(binDir, "phpstan")
	script := "#!/bin/sh\nif [ -f \"" + statePath + "\" ]; then printf 'new architecture failure\\n' >&2; else printf 'pre-existing architecture failure\\n' >&2; touch \"" + statePath + "\"; fi\nexit 1\n"
	if err := os.WriteFile(phpstanPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := NewRunner()
	baseline := runner.Baseline(t.Context(), root, RunOptions{})
	results, err := runner.RunCompared(t.Context(), root, RunOptions{}, baseline)
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected changed validation failure, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 validation result, got %d", len(results))
	}
	if results[0].Message != newValidationFailureMessage {
		t.Fatalf("unexpected validation message: %#v", results[0])
	}
	if !strings.Contains(results[0].Stderr, "new architecture failure") {
		t.Fatalf("expected changed validator stderr, got %#v", results[0])
	}
}

func TestRunnerRunsConfiguredPythonStaticChecks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[tool.ruff]\n[tool.mypy]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTool(t, root, "ruff", "#!/bin/sh\nprintf 'ruff ok\\n'\n")
	writeTool(t, root, "mypy", "#!/bin/sh\nprintf 'mypy ok\\n'\n")

	results, err := NewRunner().Run(t.Context(), root, RunOptions{})
	if err != nil {
		t.Fatalf("expected python validation to pass, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 validation results, got %d", len(results))
	}
	if results[0].Name != "ruff" || results[1].Name != "mypy" {
		t.Fatalf("unexpected validation results: %#v", results)
	}
}

func TestRunnerRunsConfiguredPythonTestsOnlyWhenRequested(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[tool.pytest.ini_options]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTool(t, root, "pytest", "#!/bin/sh\nprintf 'pytest ok\\n'\n")

	results, err := NewRunner().Run(t.Context(), root, RunOptions{})
	if err != nil {
		t.Fatalf("expected validation without tests to pass, got %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no validation results without --run-tests, got %#v", results)
	}

	results, err = NewRunner().Run(t.Context(), root, RunOptions{RunTests: true})
	if err != nil {
		t.Fatalf("expected python tests to pass, got %v", err)
	}
	if len(results) != 1 || results[0].Name != "pytest" {
		t.Fatalf("expected pytest validation result, got %#v", results)
	}
}

func writeTool(t *testing.T, root string, name string, script string) {
	t.Helper()

	binDir := filepath.Join(root, ".venv", "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}
