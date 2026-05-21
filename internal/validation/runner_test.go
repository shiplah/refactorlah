package validation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerReportsValidationFailuresAsPostApplyOnly(t *testing.T) {
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
	if results[0].Message != postApplyValidationDisclaimer {
		t.Fatalf("unexpected validation message: %#v", results[0])
	}
	if !strings.Contains(results[0].Stderr, "existing failure") {
		t.Fatalf("expected captured validator stderr, got %#v", results[0])
	}
}
