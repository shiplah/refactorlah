//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/planning"
)

func TestAnalyzerUpdatesNamespaceDeclarationAndUseStatement(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController { public const SERVICE = \\App\\Services\\Billing\\InvoiceService::class; public function service(): \\App\\Services\\Billing\\InvoiceService {} }\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Domain/Billing/InvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected 1 symbol mapping, got %#v", response.SymbolMappings)
	}
	if response.SymbolMappings[0].OldSymbol != "App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("unexpected old symbol %q", response.SymbolMappings[0].OldSymbol)
	}
	if response.SymbolMappings[0].NewSymbol != "App\\Domain\\Billing\\InvoiceService" {
		t.Fatalf("unexpected new symbol %q", response.SymbolMappings[0].NewSymbol)
	}

	assertReplacement(t, response.Replacements, "app/Services/Billing/InvoiceService.php", "App\\Services\\Billing", "App\\Domain\\Billing")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "\\App\\Services\\Billing\\InvoiceService", "\\App\\Domain\\Billing\\InvoiceService")
}

func TestAnalyzerRenamesMovedClassDeclaration(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal readonly class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController { public function service(): InvoiceService { return new InvoiceService(); } }\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Services/Billing/BillingInvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "app/Services/Billing/InvoiceService.php", "InvoiceService", "BillingInvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Services\\Billing\\BillingInvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "InvoiceService", "BillingInvoiceService")
}

func TestAnalyzerUpdatesDocblockReferences(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n/** @param iterable<InvoiceService> $services */\nfinal class InvoiceController {}\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Domain/Billing/BillingInvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "InvoiceService", "BillingInvoiceService")
}

func TestAnalyzerAddsImportsForMovedFileNamespaceLocalDependencies(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceBatch.php", "<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceBatch { public function __construct(private InvoiceFilter $range) {} }\n")
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceFilter.php", "<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceFilter {}\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Billing/Domain/InvoiceBatch.php",
			NewPath: "app/Billing/Archive/Domain/InvoiceBatch.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacementContaining(t, response.Replacements, "app/Billing/Domain/InvoiceBatch.php", "use App\\Billing\\Domain\\InvoiceFilter;")
}

func TestAnalyzerUsesComposerRootForMonorepoPaths(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "platform/composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	writeAnalyzerFixtureFile(t, root, "platform/src/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "platform/src/Http/InvoiceController.php", "<?php\nnamespace App\\Http;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController {}\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/src/Services/Billing/InvoiceService.php",
			NewPath: "platform/src/Domain/Billing/InvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}

	assertReplacement(t, response.Replacements, "platform/src/Services/Billing/InvoiceService.php", "App\\Services\\Billing", "App\\Domain\\Billing")
	assertReplacement(t, response.Replacements, "platform/src/Http/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService")
}

func writeAnalyzerFixtureFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, oldText string, newText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == newText {
			return
		}
	}
	t.Fatalf("expected replacement in %s from %q to %q, got %#v", file, oldText, newText, replacements)
}

func assertReplacementContaining(t *testing.T, replacements []adapterproto.Replacement, file string, text string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && strings.Contains(replacement.Replacement, text) {
			return
		}
	}
	t.Fatalf("expected replacement in %s containing %q, got %#v", file, text, replacements)
}
