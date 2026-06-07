package names

import "testing"

func TestSymbolParts(t *testing.T) {
	t.Parallel()

	if got := Short(`App\Billing\Domain\InvoiceBatch`); got != "InvoiceBatch" {
		t.Fatalf("expected InvoiceBatch, got %q", got)
	}
	if got := Namespace(`App\Billing\Domain\InvoiceBatch`); got != `App\Billing\Domain` {
		t.Fatalf("expected namespace, got %q", got)
	}
	if got := Short("InvoiceBatch"); got != "InvoiceBatch" {
		t.Fatalf("expected unqualified symbol, got %q", got)
	}
	if got := Namespace("InvoiceBatch"); got != "" {
		t.Fatalf("expected empty namespace, got %q", got)
	}
}

func TestContainsIdentifier(t *testing.T) {
	t.Parallel()

	content := `ComponentRenderer ComponentRendererFactory \ComponentRenderer notComponentRenderer`
	if !ContainsIdentifier(content, "ComponentRenderer") {
		t.Fatal("expected exact identifier match")
	}
	if ContainsIdentifier(content, "Renderer") {
		t.Fatal("did not expect partial identifier match")
	}
}

func TestNameBoundary(t *testing.T) {
	t.Parallel()

	text := `\App\ComponentRenderer::class`
	if IsNameBoundary(text, 0) {
		t.Fatal("backslash should not be a PHP name boundary")
	}
	if !IsNameBoundary(text, len(`\App\ComponentRenderer`)) {
		t.Fatal("colon should be a PHP name boundary")
	}
}
