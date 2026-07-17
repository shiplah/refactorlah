//go:build cgo

package php

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestSemanticHintScannerReportsVariablesPhpStringsAndTextFiles(t *testing.T) {
	root := semanticHintFixtureRoot(t, "report")

	warnings, err := SemanticHintScanner{}.Scan(root,
		[]string{"src/NodeRenderer.php"},
		[]string{"config/packages/services.yaml"},
		[]adapterproto.SymbolMapping{{
			OldSymbol: "App\\Module\\ComponentRenderer",
			NewSymbol: "App\\Module\\DirectiveRenderer",
		}},
	)
	if err != nil {
		t.Fatalf("scan semantic hints: %v", err)
	}

	assertSemanticWarning(t, warnings, "src/NodeRenderer.php", `Semantic name "componentRenderers" resembles moved symbol; consider "directiveRenderers". Not changed.`)
	assertSemanticWarning(t, warnings, "src/NodeRenderer.php", `Semantic name "component_renderer" resembles moved symbol; consider "app.directive_renderer". Not changed.`)
	assertSemanticWarning(t, warnings, "config/packages/services.yaml", `Semantic name "component_renderer" resembles moved symbol; consider "directive_renderer". Not changed.`)
}

func TestSemanticHintScannerDoesNotApplyReplacements(t *testing.T) {
	root := semanticHintFixtureRoot(t, "report-only")

	warnings, err := SemanticHintScanner{}.Scan(root,
		nil,
		[]string{"config/packages/services.yaml"},
		[]adapterproto.SymbolMapping{{
			OldSymbol: "App\\Module\\ComponentRenderer",
			NewSymbol: "App\\Module\\DirectiveRenderer",
		}},
	)
	if err != nil {
		t.Fatalf("scan semantic hints: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected report-only warning")
	}
}

func TestSemanticHintScannerSkipsNoopHintsWhenShortNameDoesNotChange(t *testing.T) {
	root := semanticHintFixtureRoot(t, "noop")

	warnings, err := SemanticHintScanner{}.Scan(root,
		[]string{"src/Consumer.php"},
		nil,
		[]adapterproto.SymbolMapping{{
			OldSymbol: "App\\Module\\Record\\Domain\\Record",
			NewSymbol: "App\\Module\\Record",
		}},
	)
	if err != nil {
		t.Fatalf("scan semantic hints: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no no-op semantic warnings, got %#v", warnings)
	}
}

func semanticHintFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-semantic-hints/"+scenario)
}

func assertSemanticWarning(t *testing.T, warnings []adapterproto.Warning, file string, message string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return
		}
	}
	t.Fatalf("missing warning %q in %#v", message, warnings)
}
