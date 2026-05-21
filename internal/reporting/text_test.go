package reporting

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTextGroupsMoveAndEditDetailsByFile(t *testing.T) {
	result := Result{
		ProjectRoot:          "/tmp/demo",
		DryRun:               true,
		AutoDetectedAdapters: []string{"php"},
		Moves: []MoveReport{
			{
				OldPath: "app/Services/Billing/InvoiceService.php",
				NewPath: "app/Domain/Billing/InvoiceService.php",
				Tracked: true,
				Mover:   "git mv",
			},
			{
				OldPath: "templates/admin/card.html.twig",
				NewPath: "templates/backoffice/card.html.twig",
				Tracked: false,
				Mover:   "filesystem rename",
			},
		},
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
		Replacements: []ReplacementReport{
			{
				File:    "app/Domain/Billing/InvoiceService.php",
				Reason:  "php-namespace-declaration",
				Adapter: "php",
				Rule:    "Refactorlah\\PhpAdapter\\Php\\Rules\\NamespaceDeclarationReplacementRule",
			},
			{
				File:    "app/Http/Controllers/InvoiceController.php",
				Reason:  "php-use-statement",
				Adapter: "php",
				Rule:    "Refactorlah\\PhpAdapter\\Php\\Rules\\UseStatementReplacementRule",
			},
			{
				File:    "app/Http/Controllers/InvoiceController.php",
				Reason:  "php-fully-qualified-class-name",
				Adapter: "php",
				Rule:    "Refactorlah\\PhpAdapter\\Php\\Rules\\FullyQualifiedClassNameReplacementRule",
			},
		},
		Warnings: []Message{{
			File:    "templates/example.twig",
			Line:    12,
			Message: "Dynamic Twig template path detected; not changed.",
		}},
		Validation: []ValidationResult{
			{
				Name:    "replacement validation",
				Message: "2 replacements validated",
			},
			{
				Name:    "composer dump-autoload",
				Message: "would run",
			},
		},
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Mode: dry",
		"Project root: /tmp/demo",
		"Semantic rewrites: php",
		"app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php [tracked, git mv]",
		"php symbol: App\\Services\\Billing\\InvoiceService -> App\\Domain\\Billing\\InvoiceService",
		"templates/admin/card.html.twig -> templates/backoffice/card.html.twig [untracked, filesystem rename]",
		"twig reference: admin/card.html.twig -> backoffice/card.html.twig",
		"app/Domain/Billing/InvoiceService.php",
		"php: namespace declaration",
		"app/Http/Controllers/InvoiceController.php",
		"php: fully qualified class reference, use statement",
		"templates/example.twig:12 Dynamic Twig template path detected; not changed.",
		"composer dump-autoload would run",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}

	for _, unexpected := range []string{
		"Adapters:",
		"PHP symbols:",
		"Workers:",
		"replacement validation",
		"Refactorlah\\PhpAdapter\\Php\\Rules\\",
	} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("did not expect %q in output:\n%s", unexpected, output)
		}
	}
}

func TestRenderTextShowsDisabledSemanticRewrites(t *testing.T) {
	result := Result{
		DryRun:           false,
		AdaptersDisabled: true,
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Mode: apply",
		"Semantic rewrites: disabled",
		"Moves:\n  (none)",
		"Edits:\n  (none)",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}
