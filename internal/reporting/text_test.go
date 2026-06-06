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
			Kind:         "twig-template",
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
		"Summary: 2 move(s), 2 edited file(s), 1 warning(s)",
		"app/Services/Billing/InvoiceService.php -> app/Domain/Billing/InvoiceService.php",
		"move: tracked, git mv",
		"php symbol: App\\Services\\Billing\\InvoiceService -> App\\Domain\\Billing\\InvoiceService",
		"templates/admin/card.html.twig -> templates/backoffice/card.html.twig",
		"template reference: admin/card.html.twig -> backoffice/card.html.twig",
		"edits (php): namespace declaration",
		"app/Http/Controllers/InvoiceController.php",
		"edits (php): fully qualified class reference, use statement",
		"templates/example.twig",
		"warning (line 12): Dynamic Twig template path detected; not changed.",
		"composer dump-autoload: would run",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}

	for _, unexpected := range []string{
		"Moves:",
		"Edits:",
		"Warnings:",
		"replacement validation",
		"Refactorlah\\PhpAdapter\\Php\\Rules\\",
	} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("did not expect %q in output:\n%s", unexpected, output)
		}
	}
}

func TestRenderTextShowsNoSemanticRewrites(t *testing.T) {
	result := Result{
		DryRun: false,
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"Mode: apply",
		"Semantic rewrites: none",
		"Summary: 0 move(s), 0 edited file(s), 0 warning(s)",
		"Files:\n  (none)",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
}

func TestRenderTextLabelsPythonModuleMappings(t *testing.T) {
	result := Result{
		DryRun:               true,
		AutoDetectedAdapters: []string{"python"},
		Moves: []MoveReport{{
			OldPath: "src/app/services/billing.py",
			NewPath: "src/app/domain/billing.py",
			Mover:   "filesystem rename",
		}},
		SymbolMappings: []SymbolMapping{{
			Kind:      "module",
			OldPath:   "src/app/services/billing.py",
			NewPath:   "src/app/domain/billing.py",
			OldSymbol: "app.services.billing",
			NewSymbol: "app.domain.billing",
		}},
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "python module: app.services.billing -> app.domain.billing") {
		t.Fatalf("expected Python module mapping label in output:\n%s", output)
	}
	if strings.Contains(output, "php symbol: app.services.billing") {
		t.Fatalf("did not expect Python mapping to be labelled as PHP:\n%s", output)
	}
}

func TestRenderTextLabelsGoPackageMappings(t *testing.T) {
	result := Result{
		DryRun:               true,
		AutoDetectedAdapters: []string{"go"},
		Moves: []MoveReport{{
			OldPath: "internal/oldpkg/service.go",
			NewPath: "internal/newpkg/service.go",
			Mover:   "filesystem rename",
		}},
		SymbolMappings: []SymbolMapping{{
			Kind:      "package",
			OldPath:   "internal/oldpkg",
			NewPath:   "internal/newpkg",
			OldSymbol: "example.com/project/internal/oldpkg",
			NewSymbol: "example.com/project/internal/newpkg",
		}, {
			Kind:      "go-type",
			OldPath:   "internal/oldpkg/old_thing.go",
			NewPath:   "internal/newpkg/new_thing.go",
			OldSymbol: "example.com/project/internal/oldpkg.OldThing",
			NewSymbol: "example.com/project/internal/newpkg.NewThing",
		}},
		PathMappings: []PathMapping{{
			Kind:         "go-import-path",
			OldPath:      "internal/oldpkg",
			NewPath:      "internal/newpkg",
			OldReference: "example.com/project/internal/oldpkg",
			NewReference: "example.com/project/internal/newpkg",
		}},
		Replacements: []ReplacementReport{
			{
				File:    "internal/consumer/use.go",
				Reason:  "go-import-path",
				Adapter: "go",
				Rule:    "go.ImportPathRule",
			},
			{
				File:    "internal/consumer/use.go",
				Reason:  "go-package-qualifier",
				Adapter: "go",
				Rule:    "go.PackageQualifierRule",
			},
			{
				File:    "internal/oldpkg/service.go",
				Reason:  "go-package-declaration",
				Adapter: "go",
				Rule:    "go.PackageDeclarationRule",
			},
			{
				File:    "internal/oldpkg/old_thing.go",
				Reason:  "go-symbol-declaration",
				Adapter: "go",
				Rule:    "go.SymbolDeclarationRule",
			},
			{
				File:    "internal/oldpkg/old_thing.go",
				Reason:  "go-local-symbol-reference",
				Adapter: "go",
				Rule:    "go.LocalSymbolReferenceRule",
			},
			{
				File:    "internal/consumer/use.go",
				Reason:  "go-imported-symbol-reference",
				Adapter: "go",
				Rule:    "go.ImportedSymbolReferenceRule",
			},
		},
	}

	var buffer bytes.Buffer
	if err := RenderText(&buffer, result); err != nil {
		t.Fatalf("render: %v", err)
	}

	output := buffer.String()
	for _, expected := range []string{
		"go package: example.com/project/internal/oldpkg -> example.com/project/internal/newpkg",
		"go type: example.com/project/internal/oldpkg.OldThing -> example.com/project/internal/newpkg.NewThing",
		"go import path: example.com/project/internal/oldpkg -> example.com/project/internal/newpkg",
		"edits (go): import path, imported symbol reference, package qualifier",
		"edits (go): local symbol reference, symbol declaration",
		"edits (go): package declaration",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected %q in output:\n%s", expected, output)
		}
	}
	for _, unexpected := range []string{
		"php symbol: example.com/project/internal/oldpkg",
		"template reference: example.com/project/internal/oldpkg",
		"go. import path",
	} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("did not expect %q in output:\n%s", unexpected, output)
		}
	}
}
