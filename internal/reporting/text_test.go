package reporting

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTextUsesCompactOneLineEntries(t *testing.T) {
	result := Result{
		ProjectRoot: "/tmp/demo",
		DryRun:      true,
		Moves: []MoveReport{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Domain/Billing/InvoiceService.php",
			Tracked: true,
			Mover:   "git mv",
		}},
		AutoDetectedAdapters: []string{"php"},
		SymbolMappings: []SymbolMapping{{
			OldPath:   "app/Services/Billing/InvoiceService.php",
			OldSymbol: "App\\Services\\Billing\\InvoiceService",
			NewSymbol: "App\\Domain\\Billing\\InvoiceService",
		}},
		PathMappings: []PathMapping{{
			OldPath:      "templates/admin/card.html.twig",
			OldReference: "admin/card.html.twig",
			NewReference: "backoffice/card.html.twig",
		}},
		EditedFiles: []EditedFile{{
			File:         "app/Http/Controllers/InvoiceController.php",
			Replacements: 2,
		}},
		ReplacementWorkerResults: []WorkerResult{{
			Worker:       "UseStatementReplacementWorker",
			Replacements: 1,
		}},
		Warnings: []Message{{
			File:    "templates/example.twig",
			Line:    12,
			Message: "Dynamic Twig template path detected; not changed.",
		}},
		Validation: []ValidationResult{{
			Name:    "replacement validation",
			Message: "2 replacements validated",
		}},
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Mode: dry-run",
		"Project root: /tmp/demo",
		"app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php [tracked, git mv]",
		"Adapters: php",
		"App\\Services\\Billing\\InvoiceService -> App\\Domain\\Billing\\InvoiceService (app/Services/Billing/InvoiceService.php)",
		"admin/card.html.twig -> backoffice/card.html.twig (templates/admin/card.html.twig)",
		"app/Http/Controllers/InvoiceController.php (2 replacement(s))",
		"UseStatementReplacementWorker: 1 replacement(s)",
		"templates/example.twig:12 Dynamic Twig template path detected; not changed.",
		"replacement validation: 2 replacements validated",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestRenderTextShowsAdaptersDisabledAndNoEditsCompactly(t *testing.T) {
	result := Result{
		DryRun:           false,
		AdaptersDisabled: true,
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Mode: apply") {
		t.Fatalf("expected apply mode in output:\n%s", output)
	}
	if !strings.Contains(output, "Adapters: (disabled)") {
		t.Fatalf("expected adapters disabled in output:\n%s", output)
	}
	if !strings.Contains(output, "Edits:\n  (none)") {
		t.Fatalf("expected empty edits in output:\n%s", output)
	}
}
