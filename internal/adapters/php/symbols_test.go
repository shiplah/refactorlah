//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
)

func TestSymbolScannerDerivesMappingForDeterministicPSR4Move(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;
final class InvoiceService {}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
	if mappings[0].OldSymbol != "App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("unexpected old symbol %q", mappings[0].OldSymbol)
	}
	if mappings[0].NewSymbol != "App\\Domain\\Billing\\InvoiceService" {
		t.Fatalf("unexpected new symbol %q", mappings[0].NewSymbol)
	}
}

func TestSymbolScannerWarnsForNonPSR4Path(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "misc/InvoiceService.php", "<?php\nfinal class InvoiceService {}\n")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "misc/InvoiceService.php",
		NewPath: "misc/MovedInvoiceService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
}

func TestSymbolScannerPrefersFilenameMatchingSymbol(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;
final class Helper {}
final class InvoiceService {}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
}

func TestSymbolScannerWarnsForAmbiguousMultipleSymbols(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;
final class A {}
final class B {}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
	}})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %#v", warnings)
	}
}

func TestSymbolScannerIgnoresNestedSymbols(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;

final class InvoiceService {}

function createInvoiceService(): object
{
    class NestedInvoiceService {}

    return new NestedInvoiceService();
}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
}

func TestSymbolScannerWarnsWhenOnlyNestedSymbolMatchesFilename(t *testing.T) {
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;

final class Helper {}

function createInvoiceService(): object
{
    final class InvoiceService {}

    return new InvoiceService();
}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
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
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", `<?php
namespace App\Services\Billing;
final class Helper {}
`)

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
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
	root := t.TempDir()
	writePHPFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {\n")

	mappings, warnings := NewSymbolScanner().Scan(root, NewPsr4Map(map[string][]string{"App\\": {"app"}}), []planning.FileMove{{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
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

func writePHPFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
