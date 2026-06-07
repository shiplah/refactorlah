package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathResolverResolveRelativePath(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()

	relative, err := resolver.Resolve(root, "app/Services/Billing")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if relative != "app/Services/Billing" {
		t.Fatalf("unexpected relative path: %s", relative)
	}
}

func TestPathResolverRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()

	if _, err := resolver.Resolve(root, "../outside"); err == nil {
		t.Fatal("expected traversal error")
	}
}

func TestPathResolverRejectsOutsideAbsolutePath(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()

	outside := filepath.Join(filepath.Dir(root), "outside")
	if _, err := resolver.Resolve(root, outside); err == nil {
		t.Fatal("expected outside-project rejection")
	}
}

func TestPathResolverNormalizesWindowsStylePath(t *testing.T) {
	root := t.TempDir()
	resolver := NewPathResolver()

	relative, err := resolver.Resolve(root, `app\Services\Billing\InvoiceService.php`)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if relative != "app/Services/Billing/InvoiceService.php" {
		t.Fatalf("unexpected normalized path: %s", relative)
	}
}

func TestPathResolverResolveMovePrefersCurrentDirectoryPath(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform")
	writePathFixture(t, filepath.Join(cwd, "src", "Old.php"))
	resolver := NewPathResolver()

	oldPath, newPath, err := resolver.ResolveMove(root, cwd, "src/Old.php", "src/New.php")
	if err != nil {
		t.Fatalf("resolve move failed: %v", err)
	}

	if oldPath != "platform/src/Old.php" || newPath != "platform/src/New.php" {
		t.Fatalf("unexpected resolved move: %s -> %s", oldPath, newPath)
	}
}

func TestPathResolverResolveMoveFallsBackToProjectRootPath(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	writePathFixture(t, filepath.Join(root, "src", "Old.php"))
	resolver := NewPathResolver()

	oldPath, newPath, err := resolver.ResolveMove(root, cwd, "src/Old.php", "src/New.php")
	if err != nil {
		t.Fatalf("resolve move failed: %v", err)
	}

	if oldPath != "src/Old.php" || newPath != "src/New.php" {
		t.Fatalf("unexpected resolved move: %s -> %s", oldPath, newPath)
	}
}

func TestPathResolverResolveMoveAcceptsParentRelativePathInsideProject(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	writePathFixture(t, filepath.Join(root, "collector", "src", "collector", "compute", "runner.py"))
	resolver := NewPathResolver()

	oldPath, newPath, err := resolver.ResolveMove(root, cwd, "../collector/src/collector/compute/runner.py", "../collector/src/collector/compute/runner_tmp.py")
	if err != nil {
		t.Fatalf("resolve move failed: %v", err)
	}

	if oldPath != "collector/src/collector/compute/runner.py" || newPath != "collector/src/collector/compute/runner_tmp.py" {
		t.Fatalf("unexpected resolved move: %s -> %s", oldPath, newPath)
	}
}

func TestPathResolverResolveMoveRejectsAmbiguousRelativePath(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform")
	writePathFixture(t, filepath.Join(root, "src", "Old.php"))
	writePathFixture(t, filepath.Join(cwd, "src", "Old.php"))
	resolver := NewPathResolver()

	_, _, err := resolver.ResolveMove(root, cwd, "src/Old.php", "src/New.php")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
}

func TestPathResolverResolveMovePrefersCurrentDirectoryWildcardMatch(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "platform")
	writePathFixture(t, filepath.Join(cwd, "src", "FooWorker.php"))
	resolver := NewPathResolver()

	oldPath, newPath, err := resolver.ResolveMove(root, cwd, "src/*Worker.php", "src/*Rule.php")
	if err != nil {
		t.Fatalf("resolve move failed: %v", err)
	}

	if oldPath != "platform/src/*Worker.php" || newPath != "platform/src/*Rule.php" {
		t.Fatalf("unexpected resolved wildcard move: %s -> %s", oldPath, newPath)
	}
}

func writePathFixture(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
}
