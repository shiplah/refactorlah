package project

import (
	"path/filepath"
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
