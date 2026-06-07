package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindComposerRootForPathsReturnsNestedProject(t *testing.T) {
	root := t.TempDir()
	platformDir := filepath.Join(root, "platform")
	if err := os.MkdirAll(platformDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(platformDir, "composer.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	composerRoot, found, err := FindComposerRootForPaths(root, []string{
		"platform/templates/billing/archive.html.twig",
		"platform/src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig",
	})
	if err != nil {
		t.Fatalf("find composer root failed: %v", err)
	}
	if !found {
		t.Fatal("expected composer root to be found")
	}
	if composerRoot != platformDir {
		t.Fatalf("expected %s, got %s", platformDir, composerRoot)
	}
}

func TestFindComposerRootsForPathsReturnsMultipleNestedProjects(t *testing.T) {
	root := t.TempDir()
	platformDir := filepath.Join(root, "platform")
	adminDir := filepath.Join(root, "admin")
	for _, dir := range []string{platformDir, adminDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	composerRoots, found, err := FindComposerRootsForPaths(root, []string{
		"platform/src/Services/Billing/InvoiceService.php",
		"admin/src/Services/User/UserService.php",
	})
	if err != nil {
		t.Fatalf("find composer roots failed: %v", err)
	}
	if !found {
		t.Fatal("expected composer roots to be found")
	}

	expected := []string{adminDir, platformDir}
	if len(composerRoots) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, composerRoots)
	}
	for index, root := range expected {
		if composerRoots[index] != root {
			t.Fatalf("expected %#v, got %#v", expected, composerRoots)
		}
	}
}
