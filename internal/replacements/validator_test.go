package replacements

import (
	"os"
	"path/filepath"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
)

func TestValidatorAcceptsValidReplacements(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app", "Foo.php")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	validator := NewValidator()
	_, err := validator.Validate(root, []adapterproto.Replacement{{
		File: "app/Foo.php", Start: 1, End: 3, Replacement: "ZZ", Reason: "test",
	}})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

func TestDeduplicateKeepsFirstIdenticalReplacement(t *testing.T) {
	replacements := Deduplicate([]adapterproto.Replacement{
		{File: "config/packages/twig.yaml", Start: 10, End: 20, Replacement: "new", Reason: "first"},
		{File: "config/packages/twig.yaml", Start: 10, End: 20, Replacement: "new", Reason: "second"},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %d", len(replacements))
	}
	if replacements[0].Reason != "first" {
		t.Fatalf("expected first replacement to win, got %q", replacements[0].Reason)
	}
}

func TestValidatorAcceptsDuplicateIdenticalReplacements(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "config", "packages", "twig.yaml")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("'App\\Old\\':\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	validator := NewValidator()
	_, err := validator.Validate(root, []adapterproto.Replacement{
		{File: "config/packages/twig.yaml", Start: 0, End: 11, Replacement: "'App\\New\\'", Reason: "yaml-twig-component-namespace"},
		{File: "config/packages/twig.yaml", Start: 0, End: 11, Replacement: "'App\\New\\'", Reason: "yaml-twig-component-namespace"},
	})
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}

func TestValidatorRejectsOverlaps(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app", "Foo.php")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	validator := NewValidator()
	_, err := validator.Validate(root, []adapterproto.Replacement{
		{File: "app/Foo.php", Start: 1, End: 3, Replacement: "ZZ", Reason: "test"},
		{File: "app/Foo.php", Start: 2, End: 4, Replacement: "YY", Reason: "test"},
	})
	if err == nil {
		t.Fatal("expected overlap error")
	}
}

func TestValidatorRejectsInvalidOffsets(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app", "Foo.php")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	validator := NewValidator()
	for _, replacements := range [][]adapterproto.Replacement{
		{{File: "app/Foo.php", Start: -1, End: 1, Replacement: "X", Reason: "test"}},
		{{File: "app/Foo.php", Start: 1, End: 8, Replacement: "X", Reason: "test"}},
		{{File: "app/Foo.php", Start: 4, End: 3, Replacement: "X", Reason: "test"}},
	} {
		if _, err := validator.Validate(root, replacements); err == nil {
			t.Fatalf("expected invalid offset error for %#v", replacements)
		}
	}
}

func TestApplierAppliesDescendingOrder(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app", "Foo.php")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("abcdef"), 0o644); err != nil {
		t.Fatal(err)
	}

	applier := NewApplier()
	err := applier.Apply(root, nil, []adapterproto.Replacement{
		{File: "app/Foo.php", Start: 4, End: 6, Replacement: "ZZ", Reason: "test"},
		{File: "app/Foo.php", Start: 0, End: 2, Replacement: "XX", Reason: "test"},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	updated, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != "XXcdZZ" {
		t.Fatalf("unexpected output: %s", string(updated))
	}
}

func TestApplierRemapsMovedFileReplacementsToNewPath(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app", "Moved.php")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	originalInfo, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	applier := NewApplier()
	err = applier.Apply(root, map[string]string{
		"app/Foo.php": "app/Moved.php",
	}, []adapterproto.Replacement{
		{File: "app/Foo.php", Start: 0, End: 2, Replacement: "XY", Reason: "test"},
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	updated, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != "XYcdef" {
		t.Fatalf("unexpected output: %s", string(updated))
	}

	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != originalInfo.Mode().Perm() {
		t.Fatalf("expected mode %#o, got %#o", originalInfo.Mode().Perm(), info.Mode().Perm())
	}
}
