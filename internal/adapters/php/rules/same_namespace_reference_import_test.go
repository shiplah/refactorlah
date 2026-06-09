//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
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

func TestSameNamespaceReferenceImportRuleInsertsClassImportsBeforeFunctionImports(t *testing.T) {
	source := []byte(`<?php
namespace App\History\Capture\Domain;

use App\Shared\Support\Collection;

use function array_reverse;
use function usort;

final readonly class CaptureCollection extends Collection
{
    public function previous(Capture $capture): ?Capture
    {
        return $capture;
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.SameNamespaceReferenceImportRule{}.Collect(document, rules.SameNamespaceReferenceImportInput{
		File:   "app/History/Capture/Domain/CaptureCollection.php",
		Source: source,
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\History\\Capture\\Domain\\Capture",
			NewSymbol: "App\\History\\Capture",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion, got %#v", replacements)
	}

	updated := string(source[:replacements[0].Start]) + replacements[0].Replacement + string(source[replacements[0].End:])
	expected := "use App\\Shared\\Support\\Collection;\nuse App\\History\\Capture;\n\nuse function array_reverse;"
	if !strings.Contains(updated, expected) {
		t.Fatalf("expected class import before function imports, got:\n%s", updated)
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
