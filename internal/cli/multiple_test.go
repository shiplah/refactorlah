package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandMoveListParsesInlinePairs(t *testing.T) {
	requests, err := ExpandMoveList([]string{
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

func TestReadMoveFileReadsPairLines(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, "moves.txt")
	if err := os.WriteFile(path, []byte("# comment\napp/Foo.php,app/Bar.php\n\ntests/A.php,tests/B.php\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	requests, err := ReadMoveFile(cwd, "moves.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
}

func TestExpandMoveListRejectsInvalidPair(t *testing.T) {
	if _, err := ExpandMoveList([]string{"app/Foo.php"}); err == nil {
		t.Fatal("expected invalid pair error")
	}
}

func TestReadMoveFileReportsInvalidLineNumber(t *testing.T) {
	cwd := t.TempDir()
	path := filepath.Join(cwd, "moves.txt")
	if err := os.WriteFile(path, []byte("app/Foo.php,app/Bar.php\ninvalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadMoveFile(cwd, "moves.txt")
	if err == nil {
		t.Fatal("expected invalid line error")
	}
	if err.Error() != "moves.txt:2: invalid move \"invalid\"; expected old-path,new-path" {
		t.Fatalf("unexpected error: %v", err)
	}
}
