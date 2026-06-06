//go:build cgo

package rules

import (
	"testing"

	"refactorlah/internal/languages/php"
)

func TestUseStatementRuleUpdatesImportedSymbol(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := UseStatementRule{}.Collect(document, UseStatementInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "App\\Domain\\Billing\\InvoiceService" {
		t.Fatalf("expected replacement symbol, got %q", replacement.Replacement)
	}
}

func TestUseStatementRulePreservesAlias(t *testing.T) {
	source := []byte("<?php\nuse App\\Services\\Billing\\InvoiceService as BillingInvoiceService;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := UseStatementRule{}.Collect(document, UseStatementInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	updated := string(source[:replacements[0].Start]) + replacements[0].Replacement + string(source[replacements[0].End:])
	expected := "<?php\nuse App\\Domain\\Billing\\InvoiceService as BillingInvoiceService;\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestUseStatementRuleDoesNotRewriteLongerSimilarImport(t *testing.T) {
	source := []byte("<?php\nuse App\\Services\\Billing\\InvoiceServiceFactory;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := UseStatementRule{}.Collect(document, UseStatementInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
