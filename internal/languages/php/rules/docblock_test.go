//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/php"
	"refactorlah/internal/languages/php/rules"
	"refactorlah/internal/replacements"
)

func TestDocblockRulesUpdateExactSymbolReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

final class Controller
{
    /**
     * @var array<string, App\Services\Billing\InvoiceService>
     * @param \App\Services\Billing\InvoiceService $service
     * @return iterable<App\Services\Billing\InvoiceService>
     * @throws App\Services\Billing\InvoiceService
     */
    public function show(): void {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	input := rules.SymbolReferenceInput{
		File:      "app/Http/Controller.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	}

	assertDocblockReplacement(t, source, rules.DocblockVarRule{}.Collect(document, input), "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService", rules.DocblockVarRuleName)
	assertDocblockReplacement(t, source, rules.DocblockParamRule{}.Collect(document, input), "\\App\\Services\\Billing\\InvoiceService", "\\App\\Domain\\Billing\\InvoiceService", rules.DocblockParamRuleName)
	assertDocblockReplacement(t, source, rules.DocblockReturnRule{}.Collect(document, input), "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService", rules.DocblockReturnRuleName)
	assertDocblockReplacement(t, source, rules.DocblockThrowsRule{}.Collect(document, input), "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService", rules.DocblockThrowsRuleName)
}

func TestDocblockRulesUpdateImportedShortReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

use App\Services\Billing\InvoiceService;

/**
 * @param iterable<InvoiceService> $services
 */
function handle(iterable $services): void {}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.DocblockParamRule{}.Collect(document, rules.SymbolReferenceInput{
		File:      "app/Http/functions.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	})

	assertDocblockReplacement(t, source, replacements, "InvoiceService", "BillingInvoiceService", rules.DocblockParamRuleName)
}

func TestDocblockRulesUpdateSameNamespaceShortReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Services\Billing;

/**
 * @return InvoiceService
 */
function make(): object {}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.DocblockReturnRule{}.Collect(document, rules.SymbolReferenceInput{
		File:         "app/Services/Billing/functions.php",
		Source:       source,
		OldSymbol:    "App\\Services\\Billing\\InvoiceService",
		NewSymbol:    "App\\Services\\Billing\\BillingInvoiceService",
		OldNamespace: "App\\Services\\Billing",
		NewNamespace: "App\\Services\\Billing",
	})

	assertDocblockReplacement(t, source, replacements, "InvoiceService", "BillingInvoiceService", rules.DocblockReturnRuleName)
}

func TestDocblockRulesRespectSymbolBoundaries(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

/**
 * @var App\Services\Billing\InvoiceServiceFactory
 * @var InvoiceServiceFactory
 */
$service = null;
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	input := rules.SymbolReferenceInput{
		File:      "app/Http/Controller.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	}

	if replacements := (rules.DocblockVarRule{}).Collect(document, input); len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func assertDocblockReplacement(t *testing.T, source []byte, replacements []replacements.Replacement, oldText string, newText string, rule string) {
	t.Helper()

	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != oldText {
			continue
		}
		if replacement.Replacement != newText {
			t.Fatalf("expected replacement %q, got %q", newText, replacement.Replacement)
		}
		if replacement.Rule != rule {
			t.Fatalf("expected rule %q, got %q", rule, replacement.Rule)
		}
		return
	}

	t.Fatalf("expected replacement from %q to %q, got %#v", oldText, newText, replacements)
}
