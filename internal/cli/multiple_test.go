package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandMultipleInputsParsesInlinePairs(t *testing.T) {
	requests, err := ExpandMultipleInputs(t.TempDir(), []string{
		"app/Foo.php,app/Bar.php",
		"tests/A.php,tests/B.php",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
	if requests[0].OldPath != "app/Foo.php" || requests[0].NewPath != "app/Bar.php" {
		t.Fatalf("unexpected first request: %#v", requests[0])
	}
}

func TestExpandMultipleInputsReadsAtFile(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, "moves.txt")
	if err := os.WriteFile(path, []byte("# comment\napp/Foo.php,app/Bar.php\n\ntests/A.php,tests/B.php\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	requests, err := ExpandMultipleInputs(cwd, []string{"@moves.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
}
