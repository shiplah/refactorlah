//go:build cgo

package rules_test

import (
	"sort"
	"testing"

	"refactorlah/internal/adapters/python"
	"refactorlah/internal/adapters/python/rules"
	"refactorlah/internal/replacements"
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

func TestImportStatementRuleUpdatesMultiImportModule(t *testing.T) {
	source := []byte("import os, app.services.billing, sys\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	})

	updated := applyPythonRuleReplacements(source, replacements)
	expected := "import os, app.domain.invoicing, sys\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestImportStatementRuleKeepsByteOffsetsStableAfterUnicodeText(t *testing.T) {
	source := []byte("# café\nimport app.services.billing\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	})

	updated := applyPythonRuleReplacements(source, replacements)
	expected := "# café\nimport app.domain.invoicing\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
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

func TestImportStatementRuleUpdatesFromParentImportName(t *testing.T) {
	source := []byte("from app.services import billing as billing_module\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	})

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %#v", replacements)
	}

	updated := string(source)
	for index := len(replacements) - 1; index >= 0; index-- {
		replacement := replacements[index]
		updated = updated[:replacement.Start] + replacement.Replacement + updated[replacement.End:]
	}
	expected := "from app.domain import invoicing as billing_module\n"
	if updated != expected {
		t.Fatalf("unexpected updated source:\n%s", updated)
	}
}

func TestImportStatementRuleUpdatesFromParentMultiImportAndAlias(t *testing.T) {
	source := []byte("from app.services import other, billing as billing_module\n")
	document, err := python.Parse(source)
	if err != nil {
		t.Fatalf("parse python: %v", err)
	}
	defer document.Close()

	replacements := rules.ImportStatementRule{}.Collect(document, rules.ImportStatementInput{
		File:      "src/app/http/controller.py",
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	})

	updated := applyPythonRuleReplacements(source, replacements)
	expected := "from app.domain import other, invoicing as billing_module\n"
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

func applyPythonRuleReplacements(source []byte, items []replacements.Replacement) string {
	sorted := append([]replacements.Replacement(nil), items...)
	sort.Slice(sorted, func(left int, right int) bool {
		return sorted[left].Start > sorted[right].Start
	})

	result := append([]byte(nil), source...)
	for _, item := range sorted {
		next := make([]byte, 0, len(result)-item.End+item.Start+len(item.Replacement))
		next = append(next, result[:item.Start]...)
		next = append(next, []byte(item.Replacement)...)
		next = append(next, result[item.End:]...)
		result = next
	}
	return string(result)
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
