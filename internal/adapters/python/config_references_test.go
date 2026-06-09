package python

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/config"
)

func TestDottedPathReferenceScannerUpdatesConfigReferences(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "pyproject.toml", "[tool.example]\nhandler = \"app.services.billing.InvoiceService\"\n# app.services.billing.CommentOnly\n")
	writePythonFixture(t, root, "config/routes.yaml", "billing_handler: app.services.billing.InvoiceService\n")

	scanConfig := config.Config{}
	replacements, err := DottedPathReferenceScanner{}.Scan(root, scan.NewIndex(root, scanConfig), []ModuleMapping{{
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	}})
	if err != nil {
		t.Fatalf("scan dotted paths: %v", err)
	}

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %#v", replacements)
	}
	assertConfigReplacement(t, replacements, "pyproject.toml")
	assertConfigReplacement(t, replacements, "config/routes.yaml")
}

func TestDottedPathReferenceScannerHonoursScanPolicy(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "config/routes.yaml", "billing_handler: app.services.billing.InvoiceService\n")

	scanConfig := config.Config{
		Exclude: []string{"config/**"},
	}
	replacements, err := DottedPathReferenceScanner{}.Scan(root, scan.NewIndex(root, scanConfig), []ModuleMapping{{
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	}})
	if err != nil {
		t.Fatalf("scan dotted paths: %v", err)
	}

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestDottedPathReferenceScannerSkipsUnsafeSuffixes(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "pyproject.toml", "handler = \"app.services.billing_extra.InvoiceService\"\n")

	scanConfig := config.Config{}
	replacements, err := DottedPathReferenceScanner{}.Scan(root, scan.NewIndex(root, scanConfig), []ModuleMapping{{
		OldModule: "app.services.billing",
		NewModule: "app.domain.invoicing",
	}})
	if err != nil {
		t.Fatalf("scan dotted paths: %v", err)
	}

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func assertConfigReplacement(t *testing.T, replacements []adapterproto.Replacement, file string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == "app.domain.invoicing" {
			return
		}
	}
	t.Fatalf("expected replacement for %s, got %#v", file, replacements)
}
