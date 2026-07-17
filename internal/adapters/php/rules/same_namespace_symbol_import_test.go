//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
)

func TestSameNamespaceSymbolImportRuleAddsImportsForMovedLocalFunctionAndConstant(t *testing.T) {
	source := []byte(`<?php
namespace App\Config;

final class Reader
{
    public function label(string $value): string
    {
        return build_label($value) . DEFAULT_LIMIT;
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceSymbolImportRule{}.Collect(document, rules.SameNamespaceSymbolImportInput{
		File:   "src/Config/Reader.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{
			{
				Kind:      "constant",
				OldSymbol: "App\\Config\\DEFAULT_LIMIT",
				NewSymbol: "App\\Shared\\DEFAULT_LIMIT",
			},
			{
				Kind:      "function",
				OldSymbol: "App\\Config\\build_label",
				NewSymbol: "App\\Shared\\build_label",
			},
		},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion, got %#v", replacements)
	}
	for _, expected := range []string{
		"use const App\\Shared\\DEFAULT_LIMIT;",
		"use function App\\Shared\\build_label;",
	} {
		if !strings.Contains(replacements[0].Replacement, expected) {
			t.Fatalf("missing %q in %#v", expected, replacements[0])
		}
	}
}

func TestSameNamespaceSymbolImportRuleSkipsMethodCalls(t *testing.T) {
	source := []byte(`<?php
namespace App\Config;

final class Reader
{
    public function label(object $formatter, string $value): string
    {
        return $formatter->build_label($value);
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceSymbolImportRule{}.Collect(document, rules.SameNamespaceSymbolImportInput{
		File:   "src/Config/Reader.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{{
			Kind:      "function",
			OldSymbol: "App\\Config\\build_label",
			NewSymbol: "App\\Shared\\build_label",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no import insertion, got %#v", replacements)
	}
}
