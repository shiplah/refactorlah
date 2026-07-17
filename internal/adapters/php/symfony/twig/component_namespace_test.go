package twig

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestComponentNamespaceScannerRewritesTwigComponentDefaults(t *testing.T) {
	root := componentNamespaceFixtureRoot(t, "defaults")

	replacements, err := ComponentNamespaceScanner{}.Scan(root, []string{"config/packages/twig_component.yaml"}, []adapterproto.SymbolMapping{{
		OldNamespace: "App\\Module\\Widget\\Ui\\Web",
		NewNamespace: "App\\Module\\Widget\\Ui\\Browser",
	}})
	if err != nil {
		t.Fatalf("scan component namespaces: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "'App\\Module\\Widget\\Ui\\Browser\\'" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestComponentNamespaceScannerSkipsNonComponentYaml(t *testing.T) {
	root := componentNamespaceFixtureRoot(t, "non-component")

	replacements, err := ComponentNamespaceScanner{}.Scan(root, []string{"config/packages/example.yaml"}, []adapterproto.SymbolMapping{{
		OldNamespace: "App\\Module\\Widget\\Ui\\Web",
		NewNamespace: "App\\Module\\Widget\\Ui\\Browser",
	}})
	if err != nil {
		t.Fatalf("scan component namespaces: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func componentNamespaceFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-symfony-twig/component-namespace/"+scenario)
}
