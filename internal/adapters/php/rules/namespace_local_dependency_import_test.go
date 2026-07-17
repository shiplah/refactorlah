//go:build cgo

package rules_test

import (
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestNamespaceLocalDependencyImportRuleAddsImportsForOldNamespaceDependencies(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/adds-imports/InvoiceBatch.php")
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
	if replacement.Start != namespaceDeclarationEnd(t, source) {
		t.Fatalf("expected import after namespace declaration, got offset %d", replacement.Start)
	}
}

func TestNamespaceLocalDependencyImportRuleSkipsDependenciesMovedToSameNamespace(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/skips-moved-same-namespace/InvoiceBatch.php")
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
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/removed-import-placement/InvoiceBatch.php")
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
	if replacement.Start != namespaceDeclarationEnd(t, source) {
		t.Fatalf("expected insertion after namespace declaration, got offset %d", replacement.Start)
	}
}

func TestNamespaceLocalDependencyImportRuleKeepsImportsThatBecomeNamespaceLocal(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/keeps-target-namespace-import/InvoiceBatch.php")
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
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/existing-vendor-import/InvoiceBatch.php")
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
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/builtin-and-declared-class/InvoiceBatch.php")
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
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/constants/SourceDocument.php")
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
	source := testfixtures.Read(t, "tests/fixtures/php-namespace-local-import-rule/constant-like-class-names/SourceDocument.php")
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

func TestNamespaceLocalDependencyImportRuleSkipsAliasQualifiedClassReferences(t *testing.T) {
	source := testfixtures.Read(t, "tests/fixtures/php-alias-qualified/before/src/Parsing/SourceDocument.php")
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
		t.Fatalf("expected alias-qualified references to remain unimported, got %#v", replacements)
	}
}

func namespaceDeclarationEnd(t *testing.T, source []byte) int {
	t.Helper()

	namespaceStart := strings.Index(string(source), "namespace ")
	if namespaceStart < 0 {
		t.Fatal("fixture is missing namespace declaration")
	}
	semicolon := strings.IndexByte(string(source[namespaceStart:]), ';')
	if semicolon < 0 {
		t.Fatal("fixture namespace declaration is missing semicolon")
	}

	return namespaceStart + semicolon + 1
}
