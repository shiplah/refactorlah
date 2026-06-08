package rules

import "testing"

func TestImportPathRuleUpdatesGoImportPath(t *testing.T) {
	source := []byte(`package php

import "example.test/acme/project/internal/parsing/treesitter"

func parse() {}
`)

	replacements, err := ImportPathRule{}.Collect(source, ImportPathInput{
		File: "internal/adapters/php/parser.go",
		Mappings: []ImportPathMapping{{
			OldImport: "example.test/acme/project/internal/parsing/treesitter",
			NewImport: "example.test/acme/project/internal/parsing/document",
		}},
	})
	if err != nil {
		t.Fatalf("collect imports: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "example.test/acme/project/internal/parsing/treesitter" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "example.test/acme/project/internal/parsing/document" {
		t.Fatalf("expected new import path, got %q", replacement.Replacement)
	}
}

func TestImportPathRuleUpdatesGroupedGoImportPath(t *testing.T) {
	source := []byte(`package php

import (
	"testing"

	"example.test/acme/project/internal/parsing/treesitter"
)
`)

	replacements, err := ImportPathRule{}.Collect(source, ImportPathInput{
		File: "internal/adapters/php/parser_test.go",
		Mappings: []ImportPathMapping{{
			OldImport: "example.test/acme/project/internal/parsing/treesitter",
			NewImport: "example.test/acme/project/internal/parsing/document",
		}},
	})
	if err != nil {
		t.Fatalf("collect imports: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	updated := string(source[:replacements[0].Start]) + replacements[0].Replacement + string(source[replacements[0].End:])
	if updated != `package php

import (
	"testing"

	"example.test/acme/project/internal/parsing/document"
)
` {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestImportPathRuleLeavesLongerSimilarPath(t *testing.T) {
	source := []byte(`package php

import "example.test/acme/project/internal/parsing/treesitterextra"
`)

	replacements, err := ImportPathRule{}.Collect(source, ImportPathInput{
		File: "internal/adapters/php/parser.go",
		Mappings: []ImportPathMapping{{
			OldImport: "example.test/acme/project/internal/parsing/treesitter",
			NewImport: "example.test/acme/project/internal/parsing/document",
		}},
	})
	if err != nil {
		t.Fatalf("collect imports: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
