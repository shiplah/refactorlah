//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/php"
	"refactorlah/internal/adapters/php/rules"
)

func TestSameNamespaceImportRemovalRuleRemovesImportsMovedToTargetNamespace(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Billing\Domain\InvoiceLineCollection;

final readonly class InvoiceBatch
{
    public function __construct(private InvoiceLineCollection $documents) {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceImportRemovalRule{}.Collect(document, rules.SameNamespaceImportRemovalInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		NewNamespace: "App\\Billing\\Archive\\Domain",
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceLineCollection",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceLineCollection",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected one removal, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "use App\\Billing\\Domain\\InvoiceLineCollection;\n\n" {
		t.Fatalf("unexpected removal range %q", string(source[replacements[0].Start:replacements[0].End]))
	}
	if replacements[0].Replacement != "" {
		t.Fatalf("expected empty replacement, got %q", replacements[0].Replacement)
	}
}

func TestSameNamespaceImportRemovalRulePreservesAliasedImports(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Billing\Domain\InvoiceLineCollection as Documents;
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceImportRemovalRule{}.Collect(document, rules.SameNamespaceImportRemovalInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		NewNamespace: "App\\Billing\\Archive\\Domain",
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceLineCollection",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceLineCollection",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
