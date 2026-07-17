package twig

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestRuleRegistryScansTwigPhpAndYamlReferences(t *testing.T) {
	root := twigScannerFixtureRoot(t, "static")

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
	root := twigScannerFixtureRoot(t, "dynamic")

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

func twigScannerFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-symfony-twig/scanner/"+scenario)
}
