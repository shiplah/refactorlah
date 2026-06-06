package rules

import "testing"

func TestPackageQualifierRuleUpdatesUnaliasedPackageReferences(t *testing.T) {
	source := []byte(`package consumer

import "example.com/project/internal/oldpkg"

func Build() {
	oldpkg.Build()
}
`)

	replacements, err := PackageQualifierRule{}.Collect(source, PackageQualifierInput{
		File: "internal/consumer/service.go",
		Mappings: []PackageQualifierMapping{{
			OldImport:  "example.com/project/internal/oldpkg",
			NewImport:  "example.com/project/internal/newpkg",
			OldPackage: "oldpkg",
			NewPackage: "newpkg",
		}},
	})
	if err != nil {
		t.Fatalf("collect package qualifiers: %v", err)
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

func TestPackageQualifierRulePreservesExplicitAliases(t *testing.T) {
	source := []byte(`package consumer

import oldpkg "example.com/project/internal/oldpkg"

func Build() {
	oldpkg.Build()
}
`)

	replacements, err := PackageQualifierRule{}.Collect(source, PackageQualifierInput{
		File: "internal/consumer/service.go",
		Mappings: []PackageQualifierMapping{{
			OldImport:  "example.com/project/internal/oldpkg",
			NewImport:  "example.com/project/internal/newpkg",
			OldPackage: "oldpkg",
			NewPackage: "newpkg",
		}},
	})
	if err != nil {
		t.Fatalf("collect package qualifiers: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestPackageQualifierRuleSkipsWhenPackageNameIsLocallyDeclared(t *testing.T) {
	source := []byte(`package consumer

import "example.com/project/internal/oldpkg"

func Build(oldpkg Builder) {
	oldpkg.Build()
}
`)

	replacements, err := PackageQualifierRule{}.Collect(source, PackageQualifierInput{
		File: "internal/consumer/service.go",
		Mappings: []PackageQualifierMapping{{
			OldImport:  "example.com/project/internal/oldpkg",
			NewImport:  "example.com/project/internal/newpkg",
			OldPackage: "oldpkg",
			NewPackage: "newpkg",
		}},
	})
	if err != nil {
		t.Fatalf("collect package qualifiers: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements for shadowed package name, got %#v", replacements)
	}
}
