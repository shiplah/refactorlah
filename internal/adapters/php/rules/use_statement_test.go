//go:build cgo

package rules_test

import (
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
)

func TestUseStatementRuleUpdatesImportedSymbol(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.UseStatementRule{}.Collect(document, rules.UseStatementInput{
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

	replacements := rules.UseStatementRule{}.Collect(document, rules.UseStatementInput{
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

func TestUseStatementRuleSkipsPlainImportsRemovedAsSameNamespace(t *testing.T) {
	source := []byte("<?php\nuse App\\Billing\\Domain\\InvoiceLineCollection;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.UseStatementRule{}.Collect(document, rules.UseStatementInput{
		File:                          "app/Billing/Domain/InvoiceBatch.php",
		OldSymbol:                     "App\\Billing\\Domain\\InvoiceLineCollection",
		NewSymbol:                     "App\\Billing\\Archive\\Domain\\InvoiceLineCollection",
		SameNamespaceRemovalNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestUseStatementRuleStillRewritesAliasesThatCannotBeRemoved(t *testing.T) {
	source := []byte("<?php\nuse App\\Billing\\Domain\\InvoiceLineCollection as Documents;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.UseStatementRule{}.Collect(document, rules.UseStatementInput{
		File:                          "app/Billing/Domain/InvoiceBatch.php",
		OldSymbol:                     "App\\Billing\\Domain\\InvoiceLineCollection",
		NewSymbol:                     "App\\Billing\\Archive\\Domain\\InvoiceLineCollection",
		SameNamespaceRemovalNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected alias import rewrite, got %#v", replacements)
	}
}

func TestUseStatementRuleDoesNotRewriteLongerSimilarImport(t *testing.T) {
	source := []byte("<?php\nuse App\\Services\\Billing\\InvoiceServiceFactory;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.UseStatementRule{}.Collect(document, rules.UseStatementInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
