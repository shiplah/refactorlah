//go:build cgo

package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"refactorlah/internal/planning"
)

func TestMoveUsesNativePHPAnalyzer(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	mustWriteFile(t, filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php"), "<?php\nnamespace App\\Services\\Billing;\nfinal class InvoiceService {}\n")
	mustWriteFile(t, filepath.Join(root, "app", "Http", "Controller.php"), "<?php\nnamespace App\\Http;\nuse App\\Services\\Billing\\InvoiceService;\nfinal class Controller {}\n")

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "app/Services/Billing/InvoiceService.php",
		NewPath: "app/Domain/Billing/InvoiceService.php",
		DryRun:  true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "php") {
		t.Fatalf("expected native php semantic source, got %#v", report.AutoDetectedAdapters)
	}
	if len(report.Replacements) == 0 {
		t.Fatalf("expected native php replacements, got %#v", report)
	}
}

func TestMoveUsesNativePythonAnalyzer(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "src", "app", "services", "billing.py"), "class InvoiceService: pass\n")
	mustWriteFile(t, filepath.Join(root, "src", "app", "http", "controller.py"), "import app.services.billing\n\nservice = app.services.billing.InvoiceService()\n")

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "src/app/services/billing.py",
		NewPath: "src/app/domain/invoicing.py",
		DryRun:  true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "python") {
		t.Fatalf("expected native python semantic source, got %#v", report.AutoDetectedAdapters)
	}
	if len(report.Replacements) == 0 {
		t.Fatalf("expected native python replacements, got %#v", report)
	}
}

func TestApplyWithNativePHPKeepsImportsBeforeDeclarations(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatch.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Domain;

use App\Billing\Archive\Domain\InvoiceLineCollection;

final readonly class InvoiceBatch
{
    public function __construct(
        public string $edition,
        public InvoiceFilter $range,
        public InvoiceTotals $stats,
        public InvoiceLineCollection $documents,
    ) {}
}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Archive", "Domain", "InvoiceLineCollection.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Archive\Domain;

final class InvoiceLineCollection {}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatchRepository.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Domain;

use App\Customer\Domain\CustomerId;

interface InvoiceBatchRepository
{
    public function changes(CustomerId $surfaceId, string $edition, InvoiceFilter $range): ?InvoiceBatch;
}
`)

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Billing/Domain/InvoiceBatch.php",
		NewPath:      "src/Billing/Archive/Domain/InvoiceBatch.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	movedFile, err := os.ReadFile(filepath.Join(root, "src", "Billing", "Archive", "Domain", "InvoiceBatch.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(movedFile), "namespace App\\Billing\\Archive\\Domain;\n\nuse App\\Billing\\Domain\\InvoiceFilter;\nuse App\\Billing\\Domain\\InvoiceTotals;\n\nfinal readonly class InvoiceBatch") {
		t.Fatalf("expected imports before moved class declaration, got:\n%s", string(movedFile))
	}

	repositoryFile, err := os.ReadFile(filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatchRepository.php"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(repositoryFile), "use App\\Customer\\Domain\\CustomerId;\nuse App\\Billing\\Archive\\Domain\\InvoiceBatch;\n\ninterface InvoiceBatchRepository") {
		t.Fatalf("expected inserted import inside import block, got:\n%s", string(repositoryFile))
	}
}

func TestApplyWithNativePythonUpdatesFixtureProject(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "python-basic"))

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/app/services/billing.py",
		NewPath:      "src/app/domain/invoicing.py",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "app", "domain", "invoicing.py")); err != nil {
		t.Fatalf("moved python file missing: %v", err)
	}

	controller := mustReadFile(t, filepath.Join(root, "src", "app", "http", "controller.py"))
	if !strings.Contains(controller, "import app.domain.invoicing") {
		t.Fatalf("expected import rewrite, got:\n%s", controller)
	}
	if !strings.Contains(controller, "from app.domain import invoicing as billing_module") {
		t.Fatalf("expected aliased parent import rewrite, got:\n%s", controller)
	}
	if !strings.Contains(controller, "def build() -> \"app.domain.invoicing.InvoiceService\":") {
		t.Fatalf("expected string annotation rewrite, got:\n%s", controller)
	}
	if !strings.Contains(controller, "literal = \"app.services.billing.InvoiceService\"") {
		t.Fatalf("expected arbitrary string to remain unchanged, got:\n%s", controller)
	}

	relativeConsumer := mustReadFile(t, filepath.Join(root, "src", "app", "services", "consumer.py"))
	if !strings.Contains(relativeConsumer, "from app.domain import invoicing") {
		t.Fatalf("expected relative import rewrite, got:\n%s", relativeConsumer)
	}
	if !strings.Contains(relativeConsumer, "return invoicing.InvoiceService()") {
		t.Fatalf("expected imported module reference rewrite, got:\n%s", relativeConsumer)
	}

	generated := mustReadFile(t, filepath.Join(root, "src", "app", "generated", "fixture.py"))
	if !strings.Contains(generated, "import app.services.billing") {
		t.Fatalf("expected configured fixture exclude to remain unchanged, got:\n%s", generated)
	}

	pyproject := mustReadFile(t, filepath.Join(root, "pyproject.toml"))
	if !strings.Contains(pyproject, `handler = "app.domain.invoicing.InvoiceService"`) {
		t.Fatalf("expected pyproject dotted path rewrite, got:\n%s", pyproject)
	}

	routes := mustReadFile(t, filepath.Join(root, "config", "routes.yaml"))
	if !strings.Contains(routes, "billing_handler: app.domain.invoicing.InvoiceService") {
		t.Fatalf("expected yaml dotted path rewrite, got:\n%s", routes)
	}
	if !strings.Contains(routes, "# app.services.billing.CommentOnly") {
		t.Fatalf("expected yaml comment to remain unchanged, got:\n%s", routes)
	}
}

func TestApplyWithNativeMixedAdaptersUpdatesBroadProject(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\":"app/"}}}`)
	mustWriteFile(t, filepath.Join(root, ".refactorlah.json"), `{"exclude":["src/app/generated/**"]}`)
	mustWriteFile(t, filepath.Join(root, "pyproject.toml"), `[tool.example]
handler = "app.services.billing.InvoiceService"
`)
	mustWriteFile(t, filepath.Join(root, "app", "Services", "Billing", "InvoiceService.php"), `<?php
namespace App\Services\Billing;

final class InvoiceService {}
`)
	mustWriteFile(t, filepath.Join(root, "app", "Http", "CheckoutController.php"), `<?php
namespace App\Http;

use App\Services\Billing\InvoiceService;

final class CheckoutController
{
    /** @param iterable<InvoiceService> $services @return \App\Services\Billing\InvoiceService */
    public function checkout(InvoiceService $service): \App\Services\Billing\InvoiceService
    {
        $template = $this->render('billing/invoice.html.twig');
        $class = \App\Services\Billing\InvoiceService::class;

        return new InvoiceService();
    }
}
`)
	mustWriteFile(t, filepath.Join(root, "app", "Fixtures", "Broken.php"), "<?php\nnamespace App\\Fixtures;\nfinal class Broken {\n")
	mustWriteFile(t, filepath.Join(root, "config", "packages", "twig.yaml"), `twig:
  default_path: '%kernel.project_dir%/templates'
  paths:
    '%kernel.project_dir%/app/Domain/Billing/Ui/Twig': Billing
`)
	mustWriteFile(t, filepath.Join(root, "config", "routes.yaml"), `billing_handler: app.services.billing.InvoiceService
template: 'billing/invoice.html.twig'
`)
	mustWriteFile(t, filepath.Join(root, "templates", "billing", "invoice.html.twig"), `{% extends 'layout.html.twig' %}
{% include 'billing/invoice.html.twig' %}
`)
	mustWriteFile(t, filepath.Join(root, "templates", "layout.html.twig"), `{% include 'billing/invoice.html.twig' %}`)
	mustWriteFile(t, filepath.Join(root, "assets", "app.js"), `import './styles/legacy.css';
console.log('ready');
`)
	mustWriteFile(t, filepath.Join(root, "assets", "styles", "legacy.css"), "body { color: black; }\n")

	mustWriteFile(t, filepath.Join(root, "src", "app", "__init__.py"), "")
	mustWriteFile(t, filepath.Join(root, "src", "app", "services", "__init__.py"), "")
	mustWriteFile(t, filepath.Join(root, "src", "app", "services", "billing.py"), `class InvoiceService:
    pass
`)
	mustWriteFile(t, filepath.Join(root, "src", "app", "http", "controller.py"), `import app.services.billing
from app.services import billing as billing_module

def build(service: "app.services.billing.InvoiceService") -> "app.services.billing.InvoiceService":
    created = app.services.billing.InvoiceService()
    return billing_module.InvoiceService()
`)
	mustWriteFile(t, filepath.Join(root, "src", "app", "services", "consumer.py"), `from . import billing

def build():
    return billing.InvoiceService()
`)
	mustWriteFile(t, filepath.Join(root, "src", "app", "generated", "fixture.py"), "import app.services.billing\n")
	mustWriteFile(t, filepath.Join(root, "src", "app", "unrelated", "broken.py"), "def broken(:\n")

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		MoveRequests: []planning.RequestedMove{
			{OldPath: "app/Services/Billing/InvoiceService.php", NewPath: "app/Domain/Billing/InvoiceProcessor.php"},
			{OldPath: "templates/billing/invoice.html.twig", NewPath: "app/Domain/Billing/Ui/Twig/invoice.html.twig"},
			{OldPath: "assets/styles/legacy.css", NewPath: "assets/styles/current.css"},
			{OldPath: "src/app/services/billing.py", NewPath: "src/app/domain/invoicing.py"},
		},
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "php") || !hasString(report.AutoDetectedAdapters, "python") {
		t.Fatalf("expected php and python semantic sources, got %#v", report.AutoDetectedAdapters)
	}

	phpMoved := mustReadFile(t, filepath.Join(root, "app", "Domain", "Billing", "InvoiceProcessor.php"))
	for _, expected := range []string{"namespace App\\Domain\\Billing;", "final class InvoiceProcessor"} {
		if !strings.Contains(phpMoved, expected) {
			t.Fatalf("expected %q in moved PHP file, got:\n%s", expected, phpMoved)
		}
	}

	controller := mustReadFile(t, filepath.Join(root, "app", "Http", "CheckoutController.php"))
	for _, expected := range []string{
		"use App\\Domain\\Billing\\InvoiceProcessor;",
		"iterable<InvoiceProcessor>",
		"\\App\\Domain\\Billing\\InvoiceProcessor",
		"\\App\\Domain\\Billing\\InvoiceProcessor::class",
		"new InvoiceProcessor()",
		"$this->render('@Billing/invoice.html.twig')",
	} {
		if !strings.Contains(controller, expected) {
			t.Fatalf("expected %q in PHP controller, got:\n%s", expected, controller)
		}
	}

	layout := mustReadFile(t, filepath.Join(root, "templates", "layout.html.twig"))
	if !strings.Contains(layout, "{% include '@Billing/invoice.html.twig' %}") {
		t.Fatalf("expected Twig include rewrite, got:\n%s", layout)
	}
	routes := mustReadFile(t, filepath.Join(root, "config", "routes.yaml"))
	for _, expected := range []string{"billing_handler: app.domain.invoicing.InvoiceService", "template: '@Billing/invoice.html.twig'"} {
		if !strings.Contains(routes, expected) {
			t.Fatalf("expected %q in routes config, got:\n%s", expected, routes)
		}
	}
	appJS := mustReadFile(t, filepath.Join(root, "assets", "app.js"))
	if !strings.Contains(appJS, "import './styles/current.css';") {
		t.Fatalf("expected static import rewrite, got:\n%s", appJS)
	}

	pythonController := mustReadFile(t, filepath.Join(root, "src", "app", "http", "controller.py"))
	for _, expected := range []string{
		"import app.domain.invoicing",
		"from app.domain import invoicing as billing_module",
		`"app.domain.invoicing.InvoiceService"`,
		"app.domain.invoicing.InvoiceService()",
		"billing_module.InvoiceService()",
	} {
		if !strings.Contains(pythonController, expected) {
			t.Fatalf("expected %q in Python controller, got:\n%s", expected, pythonController)
		}
	}
	pythonConsumer := mustReadFile(t, filepath.Join(root, "src", "app", "services", "consumer.py"))
	if !strings.Contains(pythonConsumer, "from app.domain import invoicing") || !strings.Contains(pythonConsumer, "return invoicing.InvoiceService()") {
		t.Fatalf("expected relative Python import rewrite, got:\n%s", pythonConsumer)
	}
	pyproject := mustReadFile(t, filepath.Join(root, "pyproject.toml"))
	if !strings.Contains(pyproject, `handler = "app.domain.invoicing.InvoiceService"`) {
		t.Fatalf("expected pyproject dotted path rewrite, got:\n%s", pyproject)
	}
	generated := mustReadFile(t, filepath.Join(root, "src", "app", "generated", "fixture.py"))
	if strings.Contains(generated, "app.domain.invoicing") {
		t.Fatalf("expected excluded generated file to remain unchanged, got:\n%s", generated)
	}
}
