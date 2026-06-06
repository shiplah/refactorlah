package rules

import "testing"

func TestImportedSymbolReferenceRuleUpdatesImportedSelector(t *testing.T) {
	source := []byte(`package consumer

import "example.com/project/internal/models"

func Build() models.OldThing {
	return models.OldThing{}
}
`)

	replacements, err := ImportedSymbolReferenceRule{}.Collect(source, ImportedSymbolReferenceInput{
		File: "internal/consumer/use.go",
		Mappings: []ImportedSymbolReferenceMapping{{
			OldImport:  "example.com/project/internal/models",
			OldPackage: "models",
			OldSymbol:  "OldThing",
			NewSymbol:  "NewThing",
		}},
	})
	if err != nil {
		t.Fatalf("collect imported symbol references: %v", err)
	}
	if len(replacements) != 2 {
		t.Fatalf("expected two replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != "OldThing" {
			t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
		}
		if replacement.Replacement != "NewThing" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestImportedSymbolReferenceRulePreservesImportAlias(t *testing.T) {
	source := []byte(`package consumer

import m "example.com/project/internal/models"

func Build() m.OldThing {
	return m.OldThing{}
}
`)

	replacements, err := ImportedSymbolReferenceRule{}.Collect(source, ImportedSymbolReferenceInput{
		File: "internal/consumer/use.go",
		Mappings: []ImportedSymbolReferenceMapping{{
			OldImport:  "example.com/project/internal/models",
			OldPackage: "models",
			OldSymbol:  "OldThing",
			NewSymbol:  "NewThing",
		}},
	})
	if err != nil {
		t.Fatalf("collect imported symbol references: %v", err)
	}
	if len(replacements) != 2 {
		t.Fatalf("expected two replacements for aliased import, got %#v", replacements)
	}
}
