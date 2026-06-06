//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/planning"
)

func TestAnalyzerUpdatesNamespaceDeclarationAndUseStatement(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController { public const SERVICE = \\App\\Services\\Billing\\InvoiceService::class; public function service(): \\App\\Services\\Billing\\InvoiceService {} }\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Domain/Billing/InvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected 1 symbol mapping, got %#v", response.SymbolMappings)
	}
	if response.SymbolMappings[0].OldSymbol != "App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("unexpected old symbol %q", response.SymbolMappings[0].OldSymbol)
	}
	if response.SymbolMappings[0].NewSymbol != "App\\Domain\\Billing\\InvoiceService" {
		t.Fatalf("unexpected new symbol %q", response.SymbolMappings[0].NewSymbol)
	}

	assertReplacement(t, response.Replacements, "app/Services/Billing/InvoiceService.php", "App\\Services\\Billing", "App\\Domain\\Billing")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "\\App\\Services\\Billing\\InvoiceService", "\\App\\Domain\\Billing\\InvoiceService")
}

func TestAnalyzerRenamesMovedClassDeclaration(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal readonly class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController { public function service(): InvoiceService { return new InvoiceService(); } }\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Services/Billing/BillingInvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "app/Services/Billing/InvoiceService.php", "InvoiceService", "BillingInvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Services\\Billing\\BillingInvoiceService")
	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "InvoiceService", "BillingInvoiceService")
}

func TestAnalyzerUpdatesDocblockReferences(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Http/Controllers/InvoiceController.php", "<?php\nnamespace App\\Http\\Controllers;\nuse App\\Services\\Billing\\InvoiceService;\n/** @param iterable<InvoiceService> $services */\nfinal class InvoiceController {}\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Services/Billing/InvoiceService.php",
			NewPath: "app/Domain/Billing/BillingInvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "app/Http/Controllers/InvoiceController.php", "InvoiceService", "BillingInvoiceService")
}

func TestAnalyzerAddsImportsForMovedFileNamespaceLocalDependencies(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceBatch.php", "<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceBatch { public function __construct(private InvoiceFilter $range) {} }\n")
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceFilter.php", "<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceFilter {}\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Billing/Domain/InvoiceBatch.php",
			NewPath: "app/Billing/Archive/Domain/InvoiceBatch.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacementContaining(t, response.Replacements, "app/Billing/Domain/InvoiceBatch.php", "use App\\Billing\\Domain\\InvoiceFilter;")
}

func TestAnalyzerRemovesImportsThatBecomeSameNamespace(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceBatch.php", "<?php\nnamespace App\\Billing\\Domain;\nuse App\\Billing\\Domain\\InvoiceLineCollection;\nfinal readonly class InvoiceBatch { public function __construct(private InvoiceLineCollection $documents) {} }\n")
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/InvoiceLineCollection.php", "<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceLineCollection {}\n")

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{
			{
				OldPath: "app/Billing/Domain/InvoiceBatch.php",
				NewPath: "app/Billing/Archive/Domain/InvoiceBatch.php",
			},
			{
				OldPath: "app/Billing/Domain/InvoiceLineCollection.php",
				NewPath: "app/Billing/Archive/Domain/InvoiceLineCollection.php",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "app/Billing/Domain/InvoiceBatch.php", "use App\\Billing\\Domain\\InvoiceLineCollection;", "")
	assertNoReplacement(t, response.Replacements, "app/Billing/Domain/InvoiceBatch.php", "App\\Billing\\Archive\\Domain\\InvoiceLineCollection")
}

func TestAnalyzerUpdatesTwigTemplateReferences(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "config/packages/twig.yaml", `twig:
  default_path: '%kernel.project_dir%/templates'
  paths:
    '%kernel.project_dir%/src/Billing/Archive/Listing/Ui/Web/Twig': Billing
`)
	writeAnalyzerFixtureFile(t, root, "templates/billing/archive.html.twig", `{% include 'billing/archive.html.twig' %}`)
	writeAnalyzerFixtureFile(t, root, "src/Controller.php", `<?php $this->render('billing/archive.html.twig');`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "templates/billing/archive.html.twig",
			NewPath: "src/Billing/Archive/Listing/Ui/Web/Twig/archive.html.twig",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for twig move")
	}
	if len(response.PathMappings) == 0 {
		t.Fatalf("expected twig path mappings, got %#v", response)
	}

	assertReplacement(t, response.Replacements, "templates/billing/archive.html.twig", "'billing/archive.html.twig'", "'@Billing/archive.html.twig'")
	assertReplacement(t, response.Replacements, "src/Controller.php", "'billing/archive.html.twig'", "'@Billing/archive.html.twig'")
}

func TestAnalyzerUpdatesStaticImportsForMovedAssets(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "assets/app.js", `import '../src/Billing/Archive/Listing/Ui/Web/Twig/invoice-line-preview.css';`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/Billing/Archive/Listing/Ui/Web/Twig/invoice-line-preview.css",
			NewPath: "src/Billing/Archive/InvoiceLinePreview/Ui/Web/Twig/invoice-line-preview.css",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for static asset move")
	}

	assertReplacement(t, response.Replacements, "assets/app.js", "../src/Billing/Archive/Listing/Ui/Web/Twig/invoice-line-preview.css", "../src/Billing/Archive/InvoiceLinePreview/Ui/Web/Twig/invoice-line-preview.css")
}

func TestAnalyzerUpdatesAssetMapperPathForDirectoryMove(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "config/packages/asset_mapper.yaml", `framework:
  asset_mapper:
    paths:
      - 'src/Shared/Ui/Web/'
`)

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		OldPath: "src/Shared/Ui/Web",
		NewPath: "src/Shared/Ui/Browser",
		IsDir:   true,
		Moves: []planning.FileMove{{
			OldPath: "src/Shared/Ui/Web/icon.svg",
			NewPath: "src/Shared/Ui/Browser/icon.svg",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant for project directory move")
	}

	assertReplacement(t, response.Replacements, "config/packages/asset_mapper.yaml", "'src/Shared/Ui/Web/'", "'src/Shared/Ui/Browser/'")
}

func TestAnalyzerUpdatesTwigComponentNamespaceDefaults(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Billing/Reminder/Ui/Web/ReminderComponent.php", "<?php\nnamespace App\\Billing\\Reminder\\Ui\\Web;\nfinal class ReminderComponent {}\n")
	writeAnalyzerFixtureFile(t, root, "config/packages/twig_component.yaml", `twig_component:
  defaults:
    'App\Billing\Reminder\Ui\Web\':
      template_directory: '@Billing/Reminder/Ui/Web/Twig'
`)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Billing/Reminder/Ui/Web/ReminderComponent.php",
			NewPath: "app/Billing/Archive/Reminder/Ui/Web/ReminderComponent.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertReplacement(t, response.Replacements, "config/packages/twig_component.yaml", "'App\\Billing\\Reminder\\Ui\\Web\\'", "'App\\Billing\\Archive\\Reminder\\Ui\\Web\\'")
}

func TestAnalyzerReportsSemanticRenameHints(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "composer.json", `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	writeAnalyzerFixtureFile(t, root, "app/Shared/RichText/ComponentRenderer.php", "<?php\nnamespace App\\Shared\\RichText;\ninterface ComponentRenderer {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Shared/RichText/DirectiveNodeRenderer.php", `<?php
namespace App\Shared\RichText;
final class DirectiveNodeRenderer
{
    public function __construct(private iterable $componentRenderers) {}
    public function tag(): string { return 'app.rich_text_component_renderer'; }
}
`)
	writeAnalyzerFixtureFile(t, root, "config/packages/services.yaml", `tags: ['app.rich_text_component_renderer']`)

	response, _, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "app/Shared/RichText/ComponentRenderer.php",
			NewPath: "app/Shared/RichText/DirectiveRenderer.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}

	assertWarning(t, response.Warnings, "app/Shared/RichText/DirectiveNodeRenderer.php", `Semantic name "componentRenderers" resembles moved symbol; consider "directiveRenderers". Not changed.`)
	assertWarning(t, response.Warnings, "config/packages/services.yaml", `Semantic name "component_renderer" resembles moved symbol; consider "directive_renderer". Not changed.`)
}

func TestAnalyzerUsesComposerRootForMonorepoPaths(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "platform/composer.json", `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	writeAnalyzerFixtureFile(t, root, "platform/src/Services/Billing/InvoiceService.php", "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	writeAnalyzerFixtureFile(t, root, "platform/src/Http/InvoiceController.php", "<?php\nnamespace App\\Http;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class InvoiceController {}\n")

	response, relevant, err := NewAnalyzer().Analyze(root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "platform/src/Services/Billing/InvoiceService.php",
			NewPath: "platform/src/Domain/Billing/InvoiceService.php",
		}},
	})
	if err != nil {
		t.Fatalf("analyze php: %v", err)
	}
	if !relevant {
		t.Fatal("expected php analyzer to be relevant")
	}

	assertReplacement(t, response.Replacements, "platform/src/Services/Billing/InvoiceService.php", "App\\Services\\Billing", "App\\Domain\\Billing")
	assertReplacement(t, response.Replacements, "platform/src/Http/InvoiceController.php", "App\\Services\\Billing\\InvoiceService", "App\\Domain\\Billing\\InvoiceService")
}

func writeAnalyzerFixtureFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, oldText string, newText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == newText {
			return
		}
	}
	t.Fatalf("expected replacement in %s from %q to %q, got %#v", file, oldText, newText, replacements)
}

func assertReplacementContaining(t *testing.T, replacements []adapterproto.Replacement, file string, text string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && strings.Contains(replacement.Replacement, text) {
			return
		}
	}
	t.Fatalf("expected replacement in %s containing %q, got %#v", file, text, replacements)
}

func assertNoReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, replacementText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == replacementText {
			t.Fatalf("unexpected replacement in %s to %q: %#v", file, replacementText, replacement)
		}
	}
}

func assertWarning(t *testing.T, warnings []adapterproto.Warning, file string, message string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return
		}
	}
	t.Fatalf("expected warning in %s: %s, got %#v", file, message, warnings)
}
