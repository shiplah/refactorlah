package cli

import (
	"os"
	"path/filepath"
	"testing"

	"refactorlah/internal/planning"
)

func TestExpandWildcardRequestsExpandsMatchingFiles(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "src/FooWorker.php")
	writeTestFile(t, root, "src/BarWorker.php")

	requests, err := expandWildcardRequests(root, []planning.RequestedMove{{
		OldPath: "src/*Worker.php",
		NewPath: "src/*Rule.php",
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
	if requests[0].OldPath != "src/BarWorker.php" || requests[0].NewPath != "src/BarRule.php" {
		t.Fatalf("unexpected first request: %#v", requests[0])
	}
	if requests[1].OldPath != "src/FooWorker.php" || requests[1].NewPath != "src/FooRule.php" {
		t.Fatalf("unexpected second request: %#v", requests[1])
	}
}

func TestExpandWildcardRequestsRejectsMismatchedPlaceholders(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "src/FooWorker.php")

	_, err := expandWildcardRequests(root, []planning.RequestedMove{{
		OldPath: "src/*Worker.php",
		NewPath: "src/Rule.php",
	}})
	if err == nil {
		t.Fatal("expected wildcard placeholder error")
	}
	if err.Error() != "wildcard moves require * in both old-path and new-path" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpandWildcardRequestsRejectsNoMatches(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "src/Foo.php")

	_, err := expandWildcardRequests(root, []planning.RequestedMove{{
		OldPath: "src/*Worker.php",
		NewPath: "src/*Rule.php",
	}})
	if err == nil {
		t.Fatal("expected no-match error")
	}
	if err.Error() != `wildcard path "src/*Worker.php" matched no files or directories` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeTestFile(t *testing.T, root string, relativePath string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("<?php\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
