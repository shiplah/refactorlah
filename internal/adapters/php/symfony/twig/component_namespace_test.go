package twig

import (
	"os"
	"path/filepath"
	"testing"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
)

func TestComponentNamespaceScannerRewritesTwigComponentDefaults(t *testing.T) {
	root := t.TempDir()
	writeComponentNamespaceFixture(t, root, "config/packages/twig_component.yaml", `twig_component:
  defaults:
    'App\Billing\Reminder\Ui\Web\':
      template_directory: '@Billing/Reminder/Ui/Web/Twig'
`)

	replacements, err := ComponentNamespaceScanner{}.Scan(root, []string{"config/packages/twig_component.yaml"}, []adapterproto.SymbolMapping{{
		OldNamespace: "App\\Billing\\Reminder\\Ui\\Web",
		NewNamespace: "App\\Billing\\Archive\\Reminder\\Ui\\Web",
	}})
	if err != nil {
		t.Fatalf("scan component namespaces: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "'App\\Billing\\Archive\\Reminder\\Ui\\Web\\'" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestComponentNamespaceScannerSkipsNonComponentYaml(t *testing.T) {
	root := t.TempDir()
	writeComponentNamespaceFixture(t, root, "config/packages/example.yaml", `'App\Billing\Reminder\Ui\Web\': value`)

	replacements, err := ComponentNamespaceScanner{}.Scan(root, []string{"config/packages/example.yaml"}, []adapterproto.SymbolMapping{{
		OldNamespace: "App\\Billing\\Reminder\\Ui\\Web",
		NewNamespace: "App\\Billing\\Archive\\Reminder\\Ui\\Web",
	}})
	if err != nil {
		t.Fatalf("scan component namespaces: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func writeComponentNamespaceFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
