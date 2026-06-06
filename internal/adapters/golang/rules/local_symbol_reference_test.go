package rules

import "testing"

func TestLocalSymbolReferenceRuleUpdatesResolvedSamePackageReferences(t *testing.T) {
	oldSource := []byte(`package demo

type OldThing struct{}

func (thing OldThing) Clone() OldThing {
	return OldThing{}
}
`)
	consumerSource := []byte(`package demo

func Build(value OldThing) OldThing {
	return OldThing{}
}
`)

	replacements, err := LocalSymbolReferenceRule{}.Collect(LocalSymbolReferenceInput{
		PackageName: "demo",
		Files: []GoSourceFile{
			{File: "internal/demo/old_thing.go", Source: oldSource},
			{File: "internal/demo/use.go", Source: consumerSource},
		},
		Mappings: []LocalSymbolReferenceMapping{{
			File:      "internal/demo/old_thing.go",
			OldSymbol: "OldThing",
			NewSymbol: "NewThing",
		}},
	})
	if err != nil {
		t.Fatalf("collect local symbol references: %v", err)
	}
	if len(replacements) != 6 {
		t.Fatalf("expected six resolved references, got %#v", replacements)
	}
	for _, replacement := range replacements {
		source := oldSource
		if replacement.File == "internal/demo/use.go" {
			source = consumerSource
		}
		if string(source[replacement.Start:replacement.End]) != "OldThing" {
			t.Fatalf("replacement range in %s points to %q", replacement.File, string(source[replacement.Start:replacement.End]))
		}
	}
}

func TestLocalSymbolReferenceRuleSkipsShadowedLocalIdentifier(t *testing.T) {
	oldSource := []byte(`package demo

type OldThing struct{}
`)
	consumerSource := []byte(`package demo

func Build() {
	OldThing := 1
	_ = OldThing
}
`)

	replacements, err := LocalSymbolReferenceRule{}.Collect(LocalSymbolReferenceInput{
		PackageName: "demo",
		Files: []GoSourceFile{
			{File: "internal/demo/old_thing.go", Source: oldSource},
			{File: "internal/demo/use.go", Source: consumerSource},
		},
		Mappings: []LocalSymbolReferenceMapping{{
			File:      "internal/demo/old_thing.go",
			OldSymbol: "OldThing",
			NewSymbol: "NewThing",
		}},
	})
	if err != nil {
		t.Fatalf("collect local symbol references: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected shadowed local identifier to be skipped, got %#v", replacements)
	}
}

func TestLocalSymbolReferenceRuleToleratesUnresolvedImports(t *testing.T) {
	oldSource := []byte(`package demo

type OldThing struct{}
`)
	consumerSource := []byte(`package demo

import "example.com/missing/dependency"

var _ = dependency.Value

func Build() OldThing {
	return OldThing{}
}
`)

	replacements, err := LocalSymbolReferenceRule{}.Collect(LocalSymbolReferenceInput{
		PackageName: "demo",
		Files: []GoSourceFile{
			{File: "internal/demo/old_thing.go", Source: oldSource},
			{File: "internal/demo/use.go", Source: consumerSource},
		},
		Mappings: []LocalSymbolReferenceMapping{{
			File:      "internal/demo/old_thing.go",
			OldSymbol: "OldThing",
			NewSymbol: "NewThing",
		}},
	})
	if err != nil {
		t.Fatalf("collect local symbol references: %v", err)
	}
	if len(replacements) != 2 {
		t.Fatalf("expected local references despite unresolved import, got %#v", replacements)
	}
}
