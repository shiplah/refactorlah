//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/php"
	"refactorlah/internal/adapters/php/rules"
)

func TestNamespaceDeclarationRuleUpdatesMovedFileNamespace(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceDeclarationRule{}.Collect(document, rules.NamespaceDeclarationInput{
		File:         "app/Services/Billing/InvoiceService.php",
		OldNamespace: "App\\Services\\Billing",
		NewNamespace: "App\\Domain\\Billing",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "App\\Services\\Billing" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "App\\Domain\\Billing" {
		t.Fatalf("expected replacement namespace, got %q", replacement.Replacement)
	}
	if replacement.Rule != rules.NamespaceDeclarationRuleName {
		t.Fatalf("expected rule name %q, got %q", rules.NamespaceDeclarationRuleName, replacement.Rule)
	}
}

func TestNamespaceDeclarationRuleSkipsUnchangedNamespace(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceDeclarationRule{}.Collect(document, rules.NamespaceDeclarationInput{
		File:         "app/Services/Billing/InvoiceService.php",
		OldNamespace: "App\\Services\\Billing",
		NewNamespace: "App\\Services\\Billing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestNamespaceDeclarationRuleDoesNotRewriteUseOnlyMatch(t *testing.T) {
	source := []byte("<?php\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceDeclarationRule{}.Collect(document, rules.NamespaceDeclarationInput{
		File:         "app/Http/Controllers/InvoiceController.php",
		OldNamespace: "App\\Services\\Billing",
		NewNamespace: "App\\Domain\\Billing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
