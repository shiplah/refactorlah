//go:build cgo

package rules

import (
	"testing"

	"refactorlah/internal/languages/php"
)

func TestNamespaceDeclarationRuleUpdatesMovedFileNamespace(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := NamespaceDeclarationRule{}.Collect(document, NamespaceDeclarationInput{
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
	if replacement.Rule != NamespaceDeclarationRuleName {
		t.Fatalf("expected rule name %q, got %q", NamespaceDeclarationRuleName, replacement.Rule)
	}
}

func TestNamespaceDeclarationRuleSkipsUnchangedNamespace(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := NamespaceDeclarationRule{}.Collect(document, NamespaceDeclarationInput{
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

	replacements := NamespaceDeclarationRule{}.Collect(document, NamespaceDeclarationInput{
		File:         "app/Http/Controllers/InvoiceController.php",
		OldNamespace: "App\\Services\\Billing",
		NewNamespace: "App\\Domain\\Billing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
