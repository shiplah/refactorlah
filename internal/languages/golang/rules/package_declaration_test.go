package rules

import "testing"

func TestPackageDeclarationRuleUpdatesPackageName(t *testing.T) {
	source := []byte(`package oldpkg

func Build() {}
`)

	replacements, err := PackageDeclarationRule{}.Collect(source, PackageDeclarationInput{
		File:       "internal/oldpkg/service.go",
		OldPackage: "oldpkg",
		NewPackage: "newpkg",
	})
	if err != nil {
		t.Fatalf("collect package declaration: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "oldpkg" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
	if replacements[0].Replacement != "newpkg" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestPackageDeclarationRuleSkipsDifferentPackageName(t *testing.T) {
	source := []byte("package custom\n")

	replacements, err := PackageDeclarationRule{}.Collect(source, PackageDeclarationInput{
		File:       "internal/oldpkg/service.go",
		OldPackage: "oldpkg",
		NewPackage: "newpkg",
	})
	if err != nil {
		t.Fatalf("collect package declaration: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
