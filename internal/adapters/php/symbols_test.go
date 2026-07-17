//go:build cgo

package php

import (
	"slices"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestSymbolScannerDerivesMappingForDeterministicPSR4Move(t *testing.T) {
	root := symbolFixtureRoot(t, "psr4-move")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
	if mappings[0].OldSymbol != "App\\Services\\Items\\ItemService" {
		t.Fatalf("unexpected old symbol %q", mappings[0].OldSymbol)
	}
	if mappings[0].NewSymbol != "App\\Domain\\Items\\ItemService" {
		t.Fatalf("unexpected new symbol %q", mappings[0].NewSymbol)
	}
}

func TestSymbolScannerWarnsForNonPSR4Path(t *testing.T) {
	root := symbolFixtureRoot(t, "non-psr4-path")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "misc/ItemService.php",
		NewPath: "misc/MovedItemService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
}

func TestSymbolScannerPrefersFilenameMatchingSymbol(t *testing.T) {
	root := symbolFixtureRoot(t, "prefers-filename")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
}

func TestSymbolScannerWarnsForAmbiguousMultipleSymbols(t *testing.T) {
	root := symbolFixtureRoot(t, "ambiguous-symbols")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
}

func TestSymbolScannerIgnoresNestedSymbols(t *testing.T) {
	root := symbolFixtureRoot(t, "ignores-nested-symbols")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %#v", mappings)
	}
	if mappings[0].Kind != "class" || mappings[0].OldSymbol != "App\\Services\\Items\\ItemService" {
		t.Fatalf("expected class mapping, got %#v", mappings)
	}
	if mappings[1].Kind != "function" || mappings[1].OldSymbol != "App\\Services\\Items\\createItemService" {
		t.Fatalf("expected top-level function mapping, got %#v", mappings)
	}
}

func TestSymbolScannerDerivesMappingsForTopLevelConstantsAndFunctions(t *testing.T) {
	root := symbolFixtureRoot(t, "top-level-constants-functions")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Config/symbols.php",
		NewPath: "app/Shared/symbols.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}

	got := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		got = append(got, mapping.Kind+" "+mapping.OldSymbol+" -> "+mapping.NewSymbol)
	}
	want := []string{
		"constant App\\Config\\DEFAULT_LIMIT -> App\\Shared\\DEFAULT_LIMIT",
		"constant App\\Config\\SECOND_LIMIT -> App\\Shared\\SECOND_LIMIT",
		"function App\\Config\\build_label -> App\\Shared\\build_label",
	}
	for _, expected := range want {
		if !slices.Contains(got, expected) {
			t.Fatalf("expected mapping %q, got %#v", expected, got)
		}
	}
	for _, unexpected := range []string{
		"constant App\\Config\\CLASS_LIMIT -> App\\Shared\\CLASS_LIMIT",
		"class App\\Config\\symbols -> App\\Shared\\symbols",
	} {
		if slices.Contains(got, unexpected) {
			t.Fatalf("unexpected mapping %q in %#v", unexpected, got)
		}
	}
}

func TestSymbolScannerWarnsWhenOnlyNestedSymbolMatchesFilename(t *testing.T) {
	root := symbolFixtureRoot(t, "nested-filename-match")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
	if warnings[0].Message != "Top-level symbol does not match deterministic PSR-4 filename; symbol mapping skipped." {
		t.Fatalf("unexpected warning: %#v", warnings[0])
	}
}

func TestSymbolScannerWarnsWhenSingleTopLevelSymbolDoesNotMatchFilename(t *testing.T) {
	root := symbolFixtureRoot(t, "single-symbol-mismatch")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
	if warnings[0].Message != "Top-level symbol does not match deterministic PSR-4 filename; symbol mapping skipped." {
		t.Fatalf("unexpected warning: %#v", warnings[0])
	}
}

func TestSymbolScannerWarnsWhenMovedSymbolCannotBeMapped(t *testing.T) {
	root := symbolFixtureRoot(t, "invalid-symbol")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Items/ItemService.php",
		NewPath: "app/Domain/Items/ItemService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
	if warnings[0].Message != "Moved PHP file top-level symbol could not be mapped deterministically; symbol mapping skipped." {
		t.Fatalf("unexpected warning: %#v", warnings[0])
	}
}

func symbolFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-symbols/"+scenario)
}
