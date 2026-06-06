//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/python"
	"refactorlah/internal/languages/python/rules"
)

func TestImportStatementRuleUpdatesImportModule(t *testing.T) {
	source := []byte("import app.services.billing as billing\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.billing",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "app.services.billing" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "app.domain.billing" {
		t.Fatalf("expected replacement module, got %q", replacement.Replacement)
	}
}

func TestImportStatementRuleUpdatesFromImportModule(t *testing.T) {
	source := []byte("from app.services.billing import invoice_service\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.billing",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	updated := string(source[:replacements[0].Start]) + replacements[0].Replacement + string(source[replacements[0].End:])
	expected := "from app.domain.billing import invoice_service\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestImportStatementRuleDoesNotRewriteLongerSimilarModule(t *testing.T) {
	source := []byte("import app.services.billing_extra\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.billing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestImportStatementRuleLeavesRelativeImportForDedicatedRule(t *testing.T) {
	source := []byte("from .billing import invoice_service\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/services/controller.py",
		OldModule: "billing",
		NewModule: "domain.billing",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
