package replacements

import (
	"os"
	"path/filepath"
	"testing"

	adapterproto "refactorlah/internal/adapters"
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
