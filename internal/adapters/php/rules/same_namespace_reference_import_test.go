//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/php"
	"github.com/NickSdot/refactorlah/internal/adapters/php/rules"
)

func TestSameNamespaceReferenceImportRuleAddsImportForMovedLocalReference(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Customer\Domain\CustomerId;

interface InvoiceBatchRepository
{
    public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceReferenceImportRule{}.Collect(document, rules.SameNamespaceReferenceImportInput{
		File:   "app/Billing/Domain/InvoiceBatchRepository.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceBatch",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceBatch",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion, got %#v", replacements)
	}
	if !strings.Contains(replacements[0].Replacement, "use App\\Billing\\Archive\\Domain\\InvoiceBatch;") {
		t.Fatalf("missing moved symbol import: %#v", replacements[0])
	}
}

func TestSameNamespaceReferenceImportRuleSkipsReferencesResolvedByExistingImport(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use Vendor\InvoiceBatch;

final class Consumer
{
    public function changes(): ?InvoiceBatch {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceReferenceImportRule{}.Collect(document, rules.SameNamespaceReferenceImportInput{
		File:   "app/Billing/Domain/Consumer.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceBatch",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceBatch",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected imported reference to remain unchanged, got %#v", replacements)
	}
}

func TestSameNamespaceReferenceImportRuleSkipsExistingNewImport(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceBatch;

final class Consumer
{
    public function changes(): ?InvoiceBatch {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceReferenceImportRule{}.Collect(document, rules.SameNamespaceReferenceImportInput{
		File:   "app/Billing/Domain/Consumer.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceBatch",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceBatch",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected existing import to be enough, got %#v", replacements)
	}
}
