//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"testing"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
)

func TestSemanticHintScannerReportsVariablesPhpStringsAndTextFiles(t *testing.T) {
	root := t.TempDir()
	writeSemanticHintFixture(t, root, "src/DirectiveNodeRenderer.php", `<?php
final class DirectiveNodeRenderer
{
    public function __construct(private iterable $componentRenderers) {}

    public function tag(): string
    {
        return 'app.rich_text_component_renderer';
    }
}
`)
	writeSemanticHintFixture(t, root, "config/packages/services.yaml", `services:
  tags: ['app.rich_text_component_renderer']
`)

	warnings, err := SemanticHintScanner{}.Scan(root,
		[]string{"src/DirectiveNodeRenderer.php"},
		[]string{"config/packages/services.yaml"},
		[]adapterproto.SymbolMapping{{
			OldSymbol: "App\\Shared\\RichText\\ComponentRenderer",
			NewSymbol: "App\\Shared\\RichText\\DirectiveRenderer",
		}},
	)
	if err != nil {
		t.Fatalf("scan semantic hints: %v", err)
	}

	assertSemanticWarning(t, warnings, "src/DirectiveNodeRenderer.php", `Semantic name "componentRenderers" resembles moved symbol; consider "directiveRenderers". Not changed.`)
	assertSemanticWarning(t, warnings, "src/DirectiveNodeRenderer.php", `Semantic name "component_renderer" resembles moved symbol; consider "app.rich_text_directive_renderer". Not changed.`)
	assertSemanticWarning(t, warnings, "config/packages/services.yaml", `Semantic name "component_renderer" resembles moved symbol; consider "directive_renderer". Not changed.`)
}

func TestSemanticHintScannerDoesNotApplyReplacements(t *testing.T) {
	root := t.TempDir()
	writeSemanticHintFixture(t, root, "config/packages/services.yaml", `tags: ['app.rich_text_component_renderer']`)

	warnings, err := SemanticHintScanner{}.Scan(root,
		nil,
		[]string{"config/packages/services.yaml"},
		[]adapterproto.SymbolMapping{{
			OldSymbol: "App\\Shared\\RichText\\ComponentRenderer",
			NewSymbol: "App\\Shared\\RichText\\DirectiveRenderer",
		}},
	)
	if err != nil {
		t.Fatalf("scan semantic hints: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected report-only warning")
	}
}

func writeSemanticHintFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
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
