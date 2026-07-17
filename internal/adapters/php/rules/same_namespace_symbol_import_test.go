//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestSameNamespaceSymbolImportRuleAddsImportsForMovedLocalFunctionAndConstant(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-unqualified-symbols/before/src/Config/Reader.php")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceSymbolImportRule{}.Collect(document, rules.SameNamespaceSymbolImportInput{
		File:   "src/Config/Reader.php",
		Source: source,
		Mappings: []adapterproto.SymbolMapping{
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
	source := testfixtures.Read(t, "tests/fixtures/php-unqualified-symbols/rule/method-call/Reader.php")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceSymbolImportRule{}.Collect(document, rules.SameNamespaceSymbolImportInput{
		File:   "src/Config/Reader.php",
		Source: source,
		Mappings: []adapterproto.SymbolMapping{{
			Kind:      "function",
			OldSymbol: "App\\Config\\build_label",
			NewSymbol: "App\\Shared\\build_label",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no import insertion, got %#v", replacements)
	}
}

func TestSameNamespaceSymbolImportRuleWarnsForMovedSymbolsWithoutComposerFiles(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-unqualified-symbols/no-composer-files/before/src/Config/Reader.php")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	warnings := rules.SameNamespaceSymbolImportRule{}.CollectWarnings(document, rules.SameNamespaceSymbolImportInput{
		File:   "src/Config/Reader.php",
		Source: source,
		Mappings: []adapterproto.SymbolMapping{
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

	if len(warnings) != 2 {
		t.Fatalf("expected two warnings, got %#v", warnings)
	}
	for _, warning := range warnings {
		if warning.File != "src/Config/Reader.php" || warning.Line == 0 {
			t.Fatalf("unexpected warning location: %#v", warning)
		}
		if !strings.Contains(warning.Message, "not Composer autoload.files") {
			t.Fatalf("unexpected warning message: %#v", warning)
		}
	}
}
