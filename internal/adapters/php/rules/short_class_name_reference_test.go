//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/php"
	"refactorlah/internal/adapters/php/rules"
)

func TestShortClassNameReferenceRuleUpdatesImportedReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

use App\Services\Billing\InvoiceService;

final class InvoiceController implements InvoiceService
{
    public function __construct(private InvoiceService $service) {}

    public function show(?InvoiceService $service): InvoiceService
    {
        if (!$service instanceof InvoiceService) {
            return new InvoiceService();
        }

        return InvoiceService::make();
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Http/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 7 {
		t.Fatalf("expected 7 replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != "InvoiceService" {
			t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
		}
		if replacement.Replacement != "BillingInvoiceService" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestShortClassNameReferenceRuleUpdatesAttributeAndTypePositions(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

use App\Services\Billing\InvoiceService;

#[Attr(service: InvoiceService::class)]
final class InvoiceController
{
    private InvoiceService $service;

    public function show(InvoiceService $service): InvoiceService
    {
        return $service;
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Http/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 4 {
		t.Fatalf("expected 4 replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != "InvoiceService" {
			t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
		}
		if replacement.Replacement != "BillingInvoiceService" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestShortClassNameReferenceRuleUpdatesSameNamespaceReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Services\Billing;

final class InvoiceController
{
    private ?InvoiceService $service = null;

    public function show(InvoiceService $service): InvoiceService
    {
        if (!$service instanceof InvoiceService) {
            return new InvoiceService();
        }

        return InvoiceService::make();
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Services/Billing/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Services\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 6 {
		t.Fatalf("expected 6 replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != "InvoiceService" {
			t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
		}
		if replacement.Replacement != "BillingInvoiceService" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestShortClassNameReferenceRuleUpdatesTraitUseReferences(t *testing.T) {
	source := []byte(`<?php
namespace App\Http;

use App\Services\Billing\BuildsInvoice;

final class InvoiceController
{
    use BuildsInvoice;
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Http/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\BuildsInvoice",
		NewSymbol: "App\\Domain\\Billing\\BuildsBillingInvoice",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected one trait-use replacement, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "BuildsInvoice" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
	if replacements[0].Replacement != "BuildsBillingInvoice" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestShortClassNameReferenceRuleSkipsSameNamespaceReferencesShadowedByImport(t *testing.T) {
	source := []byte(`<?php
namespace App\Services\Billing;

use Vendor\InvoiceService;

$service = new InvoiceService();
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Services/Billing/Factory.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Services\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestShortClassNameReferenceRuleRequiresPlainImport(t *testing.T) {
	source := []byte(`<?php
use App\Services\Billing\InvoiceService as Billing;

$service = new Billing();
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Http/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestShortClassNameReferenceRuleDoesNotRewriteVariablesOrDeclarations(t *testing.T) {
	source := []byte(`<?php
use App\Services\Billing\InvoiceService;

final class InvoiceService {}

$InvoiceService = 'not a type';
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ShortClassNameReferenceRule{}.Collect(document, rules.ShortClassNameReferenceInput{
		File:      "app/Http/InvoiceController.php",
		Source:    source,
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\BillingInvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
