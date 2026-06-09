package twig

import (
	"os"
	"path/filepath"
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
)

func TestRuleRegistryScansTwigPhpAndYamlReferences(t *testing.T) {
	root := t.TempDir()
	writeScannerFixture(t, root, "templates/page.html.twig", `{% include 'admin/card.html.twig' %}`)
	writeScannerFixture(t, root, "src/Controller.php", `<?php $this->render('admin/card.html.twig');`)
	writeScannerFixture(t, root, "config/packages/card.yaml", `template: 'admin/card.html.twig'`)

	replacements, warnings, err := NewRuleRegistry().Scan(root,
		[]string{"src/Controller.php", "config/packages/card.yaml"},
		[]string{"templates/page.html.twig"},
		[]adapterproto.PathMapping{{
			OldReference: "admin/card.html.twig",
			NewReference: "backoffice/card.html.twig",
		}},
	)
	if err != nil {
		t.Fatalf("scan twig references: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(replacements) != 3 {
		t.Fatalf("expected three replacements, got %#v", replacements)
	}
}

func TestRuleRegistryWarnsForDynamicTemplateReferences(t *testing.T) {
	root := t.TempDir()
	writeScannerFixture(t, root, "templates/page.html.twig", `{% include admin/card.html.twig %}`)
	writeScannerFixture(t, root, "src/Controller.php", `<?php $this->render($section.'/admin/card.html.twig');`)

	_, warnings, err := NewRuleRegistry().Scan(root,
		[]string{"src/Controller.php"},
		[]string{"templates/page.html.twig"},
		[]adapterproto.PathMapping{{
			OldReference: "admin/card.html.twig",
			NewReference: "backoffice/card.html.twig",
		}},
	)
	if err != nil {
		t.Fatalf("scan twig references: %v", err)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected two warnings, got %#v", warnings)
	}
}

func writeScannerFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
