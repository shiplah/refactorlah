package rules

import "testing"

func TestSymbolDeclarationRuleUpdatesTopLevelType(t *testing.T) {
	source := []byte(`package demo

type OldThing struct{}
`)

	replacements, err := SymbolDeclarationRule{}.Collect(source, SymbolDeclarationInput{
		File: "internal/demo/old_thing.go",
		Mappings: []SymbolDeclarationMapping{{
			File:      "internal/demo/old_thing.go",
			OldSymbol: "OldThing",
			NewSymbol: "NewThing",
			Kind:      "type",
		}},
	})
	if err != nil {
		t.Fatalf("collect symbol declaration: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "OldThing" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
	if replacements[0].Replacement != "NewThing" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestSymbolDeclarationRuleUpdatesTopLevelFunction(t *testing.T) {
	source := []byte(`package demo

func oldThing() {}
`)

	replacements, err := SymbolDeclarationRule{}.Collect(source, SymbolDeclarationInput{
		File: "internal/demo/old_thing.go",
		Mappings: []SymbolDeclarationMapping{{
			File:      "internal/demo/old_thing.go",
			OldSymbol: "oldThing",
			NewSymbol: "newThing",
			Kind:      "function",
		}},
	})
	if err != nil {
		t.Fatalf("collect symbol declaration: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "oldThing" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
}
