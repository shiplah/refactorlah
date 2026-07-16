//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
)

func TestNamespaceLocalDependencyImportRuleAddsImportsForOldNamespaceDependencies(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function __construct(private InvoiceFilter $range) {}

    public function stats(): InvoiceTotals
    {
        return new InvoiceTotals();
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion only, got %#v", replacements)
	}

	replacement := replacements[0]
	if replacement.Reason != "php-namespace-local-import" {
		t.Fatalf("expected import insertion, got %#v", replacement)
	}
	if !strings.Contains(replacement.Replacement, "use App\\Billing\\Domain\\InvoiceFilter;") {
		t.Fatalf("missing InvoiceFilter import in %q", replacement.Replacement)
	}
	if !strings.Contains(replacement.Replacement, "use App\\Billing\\Domain\\InvoiceTotals;") {
		t.Fatalf("missing InvoiceTotals import in %q", replacement.Replacement)
	}
	if replacement.Start != len("<?php\nnamespace App\\Billing\\Domain;") {
		t.Fatalf("expected import after namespace declaration, got offset %d", replacement.Start)
	}
}

func TestNamespaceLocalDependencyImportRuleSkipsDependenciesMovedToSameNamespace(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function stats(): InvoiceTotals
    {
        return new InvoiceTotals();
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceTotals",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceTotals",
		}},
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestNamespaceLocalDependencyImportRuleDoesNotInsertAfterRemovedImports(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceLineCollection;

final readonly class InvoiceBatch
{
    public function __construct(
        private InvoiceFilter $range,
        private InvoiceLineCollection $documents,
    ) {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
		Mappings: []rules.SymbolMappingReference{{
			OldSymbol: "App\\Billing\\Domain\\InvoiceLineCollection",
			NewSymbol: "App\\Billing\\Archive\\Domain\\InvoiceLineCollection",
		}},
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion only, got %#v", replacements)
	}
	replacement := replacements[0]
	if !strings.Contains(replacement.Replacement, "use App\\Billing\\Domain\\InvoiceFilter;") {
		t.Fatalf("missing InvoiceFilter import in %q", replacement.Replacement)
	}
	if replacement.Start != len("<?php\nnamespace App\\Billing\\Domain;") {
		t.Fatalf("expected insertion after namespace declaration, got offset %d", replacement.Start)
	}
}

func TestNamespaceLocalDependencyImportRuleKeepsImportsThatBecomeNamespaceLocal(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceLineCollection;

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

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected target-namespace import to remain short, got %#v", replacements)
	}
}

func TestNamespaceLocalDependencyImportRuleSkipsReferencesResolvedByExistingImport(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

use Vendor\InvoiceFilter;

final readonly class InvoiceBatch
{
    public function __construct(private InvoiceFilter $range) {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected imported reference to remain unchanged, got %#v", replacements)
	}
}

func TestNamespaceLocalDependencyImportRuleSkipsBuiltinsAndDeclaredClass(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

final readonly class InvoiceBatch
{
    public function __construct(private string $name) {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "app/Billing/Domain/InvoiceBatch.php",
		Source:       source,
		OldNamespace: "App\\Billing\\Domain",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestNamespaceLocalDependencyImportRuleSkipsGlobalAndMagicConstants(t *testing.T) {
	source := []byte(`<?php
namespace App\Parsing;

use FilesystemIterator;
use RuntimeException;

final readonly class SourceDocument
{
    private const LABELS = ['section'];

    public static function from(string $contents): self
    {
        if (! preg_match_all('/section/', $contents, $matches, PREG_OFFSET_CAPTURE)) {
            throw new RuntimeException('Missing section.');
        }

        if (FilesystemIterator::SKIP_DOTS === 0 || in_array('section', self::LABELS, true)) {
            throw new RuntimeException('Invalid section.');
        }

        return new self(dirname(__DIR__, 2));
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "src/Parsing/SourceDocument.php",
		Source:       source,
		OldNamespace: "App\\Parsing",
		NewNamespace: "App\\Analysis\\Parsing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected global and magic constants to remain unimported, got %#v", replacements)
	}
}

func TestNamespaceLocalDependencyImportRuleKeepsConstantLikeClassNamesInTypePositions(t *testing.T) {
	source := []byte(`<?php
namespace App\Parsing;

final readonly class SourceDocument
{
    public function __construct(
        private XML_READER $reader,
        private __TOKEN__ $token,
    ) {}
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.NamespaceLocalDependencyImportRule{}.Collect(document, rules.NamespaceLocalDependencyImportInput{
		File:         "src/Parsing/SourceDocument.php",
		Source:       source,
		OldNamespace: "App\\Parsing",
		NewNamespace: "App\\Analysis\\Parsing",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected import insertion only, got %#v", replacements)
	}
	if !strings.Contains(replacements[0].Replacement, "use App\\Parsing\\XML_READER;") {
		t.Fatalf("missing XML_READER import in %q", replacements[0].Replacement)
	}
	if !strings.Contains(replacements[0].Replacement, "use App\\Parsing\\__TOKEN__;") {
		t.Fatalf("missing __TOKEN__ import in %q", replacements[0].Replacement)
	}
}
