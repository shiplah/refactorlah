//go:build cgo

package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
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

func TestApplyWithNativePHPUpdatesCaptureMoveImportsAndKeepsFunctionImportGroup(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "platform", "composer.json"), `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	mustWriteFile(t, filepath.Join(root, "platform", "src", "History", "Capture", "Domain", "Capture.php"), `<?php
namespace App\History\Capture\Domain;

final readonly class Capture
{
    public function __construct(
        public int $capturedAt,
        public string $captureKey,
    ) {}
}
`)
	mustWriteFile(t, filepath.Join(root, "platform", "src", "History", "Capture", "Domain", "CaptureCollection.php"), `<?php
namespace App\History\Capture\Domain;

use App\Shared\Support\Collection;

use function array_reverse;
use function usort;

final readonly class CaptureCollection extends Collection
{
    public function previous(Capture $capture): ?Capture
    {
        return $capture;
    }
}
`)
	mustWriteFile(t, filepath.Join(root, "platform", "src", "History", "ComparisonDocument", "Application", "DocumentPageDataMapper.php"), `<?php
namespace App\History\ComparisonDocument\Application;

use App\History\Capture\Domain\Capture;
use App\History\Capture\Domain\CaptureCollection;

final readonly class DocumentPageDataMapper
{
    public function map(?object $artifacts): CaptureCollection
    {
        $artifacts ?? throw new \LogicException('Rendered artifacts are required.');

        new ComparisonCaptures(
            old: new Capture(
                capturedAt: 1_779_194_233,
                captureKey: $artifacts?->olderCaptureKey,
            ),
            new: new Capture(
                capturedAt: 1_779_448_907,
                captureKey: $artifacts?->newerCaptureKey,
            ),
        );

        return new CaptureCollection();
    }
}
`)
	mustWriteFile(t, filepath.Join(root, "platform", "tests", "History", "ComparisonDocument", "Ui", "Web", "UnifiedPatchRendererTest.php"), `<?php
namespace App\Tests\History\ComparisonDocument\Ui\Web;

use App\History\Capture\Domain\Capture;

final class UnifiedPatchRendererTest
{
    #[Test]
    public function itRenders(): void
    {
        new Capture(1_779_194_233, '2026-05-19-1237');
    }
}
`)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "platform/src/History/Capture/Domain/Capture.php",
		NewPath:      "platform/src/History/Capture.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	applicationFile := mustReadFile(t, filepath.Join(root, "platform", "src", "History", "ComparisonDocument", "Application", "DocumentPageDataMapper.php"))
	testFile := mustReadFile(t, filepath.Join(root, "platform", "tests", "History", "ComparisonDocument", "Ui", "Web", "UnifiedPatchRendererTest.php"))
	for file, content := range map[string]string{
		"application": applicationFile,
		"test":        testFile,
	} {
		if strings.Contains(content, "use App\\History\\Capture\\Domain\\Capture;") {
			t.Fatalf("expected stale Capture import to be rewritten in %s file, got:\n%s", file, content)
		}
		if !strings.Contains(content, "use App\\History\\Capture;") {
			t.Fatalf("expected new Capture import in %s file, got:\n%s", file, content)
		}
	}

	collectionFile := mustReadFile(t, filepath.Join(root, "platform", "src", "History", "Capture", "Domain", "CaptureCollection.php"))
	expectedImportBlock := "use App\\Shared\\Support\\Collection;\nuse App\\History\\Capture;\n\nuse function array_reverse;"
	if !strings.Contains(collectionFile, expectedImportBlock) {
		t.Fatalf("expected class import before function imports, got:\n%s", collectionFile)
	}
}

func TestApplyWithNativePHPFailsValidationWhenStaleOldSymbolSurvives(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	mustWriteFile(t, filepath.Join(root, "src", "Example", "Old", "Thing.php"), `<?php
namespace App\Example\Old;

final class Thing {}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Example", "Consumer.php"), `<?php
?>
App\Example\Old\Thing
`)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "src/Example/Old/Thing.php",
		NewPath: "src/Example/New/Thing.php",
		Apply:   true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitValidationFailed {
		t.Fatalf("expected validation failure, got %d %#v", exitCode, report.Errors)
	}
	if !hasValidation(report.Validation, "stale PHP symbol scan", "failed") {
		t.Fatalf("expected stale PHP symbol scan failure, got %#v", report.Validation)
	}
	if len(report.Validation) == 0 || !strings.Contains(report.Validation[len(report.Validation)-1].Stdout, "src/Example/Consumer.php:3 App\\Example\\Old\\Thing") {
		t.Fatalf("expected stale symbol location in validation output, got %#v", report.Validation)
	}
}

func TestApplyWithNativePHPRoundTripDirectoryMoveKeepsImportedTypeHints(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "composer.json"), `{"autoload":{"psr-4":{"App\\":"src/"}}}`)
	mustWriteFile(t, filepath.Join(root, "src", "Schema", "Model", "InvoiceReminder.php"), `<?php

declare(strict_types=1);

namespace App\Schema\Model;

final class InvoiceReminder {}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Reminder", "Domain", "ReminderMessage.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Reminder\Domain;

final class ReminderMessage {}
`)
	mustWriteFile(t, filepath.Join(root, "src", "Billing", "Reminder", "Application", "InvoiceReminderMapper.php"), `<?php

declare(strict_types=1);

namespace App\Billing\Reminder\Application;

use App\Schema\Model\InvoiceReminder;
use App\Billing\Reminder\Domain\ReminderMessage;

final readonly class InvoiceReminderMapper
{
    public function map(InvoiceReminder $notice): ReminderMessage
    {
        return new ReminderMessage();
    }
}
`)

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Billing/Reminder/Application",
		NewPath:      "src/Billing/Reminder/Mapper",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("first move failed: %d %#v", exitCode, report.Errors)
	}

	report, exitCode = command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Billing/Reminder/Mapper",
		NewPath:      "src/Billing/Reminder/Application",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("second move failed: %d %#v", exitCode, report.Errors)
	}

	movedBack := mustReadFile(t, filepath.Join(root, "src", "Billing", "Reminder", "Application", "InvoiceReminderMapper.php"))
	for _, expected := range []string{
		"namespace App\\Billing\\Reminder\\Application;",
		"use App\\Schema\\Model\\InvoiceReminder;",
		"public function map(InvoiceReminder $notice): ReminderMessage",
	} {
		if !strings.Contains(movedBack, expected) {
			t.Fatalf("expected %q after round trip, got:\n%s", expected, movedBack)
		}
	}
	if strings.Contains(movedBack, "\\App\\Billing\\Reminder\\Application\\\\App\\Billing\\Reminder\\Mapper") {
		t.Fatalf("round trip produced duplicated namespace reference:\n%s", movedBack)
	}
}

func TestApplyWithNativePHPMoveKeepsAliasQualifiedReferencesUnimported(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-alias-qualified"))

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Parsing",
		NewPath:      "src/Analysis/Parsing",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	movedFile := mustReadFile(t, filepath.Join(root, "src", "Analysis", "Parsing", "SourceDocument.php"))
	for _, unexpected := range []string{
		"use App\\Parsing\\Catch_;",
		"use App\\Parsing\\Variable;",
	} {
		if strings.Contains(movedFile, unexpected) {
			t.Fatalf("expected no alias-segment import %q, got:\n%s", unexpected, movedFile)
		}
	}
	for _, expected := range []string{
		"namespace App\\Analysis\\Parsing;",
		"use External\\Syntax\\Expr;",
		"use External\\Syntax\\Stmt;",
		"public function variable(Stmt\\Catch_ $catch): ?Expr\\Variable",
		"instanceof Expr\\Variable",
	} {
		if !strings.Contains(movedFile, expected) {
			t.Fatalf("expected %q in moved file, got:\n%s", expected, movedFile)
		}
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

func TestApplyWithNativeFixtureCorpusUpdatesAllAdapters(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "native-mixed"))
	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		MoveRequests: []planning.RequestedMove{
			{OldPath: "app/Services/Billing/InvoiceService.php", NewPath: "app/Domain/Billing/InvoiceProcessor.php"},
			{OldPath: "templates/billing/invoice.html.twig", NewPath: "app/Domain/Billing/Ui/Twig/invoice.html.twig"},
			{OldPath: "assets/styles/legacy.css", NewPath: "assets/styles/current.css"},
			{OldPath: "src/app/services/billing.py", NewPath: "src/app/domain/invoicing.py"},
			{OldPath: "internal/oldpkg/old_service.go", NewPath: "internal/newpkg/new_service.go"},
			{OldPath: "internal/oldpkg/old_worker.go", NewPath: "internal/newpkg/new_worker.go"},
			{OldPath: "internal/oldpkg/helper.go", NewPath: "internal/newpkg/helper.go"},
			{OldPath: "internal/oldpkg/service_test.go", NewPath: "internal/newpkg/service_test.go"},
		},
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	for _, expected := range []string{"php", "python", "go"} {
		if !hasString(report.AutoDetectedAdapters, expected) {
			t.Fatalf("expected %s semantic source, got %#v", expected, report.AutoDetectedAdapters)
		}
	}
	for _, warning := range report.Warnings {
		if strings.Contains(warning.Message, "could not be parsed") ||
			strings.Contains(warning.Message, "not analysed") ||
			strings.Contains(warning.Message, "parser reported") {
			t.Fatalf("expected warnings to avoid parser details, got %#v", warning)
		}
		for _, brokenFile := range []string{"app/Fixtures/Broken.php", "internal/unrelated/broken.go", "src/app/unrelated/broken.py"} {
			if warning.File == brokenFile {
				t.Fatalf("expected unrelated broken file %s to avoid warnings, got %#v", brokenFile, warning)
			}
		}
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

	service := mustReadFile(t, filepath.Join(root, "internal", "newpkg", "new_service.go"))
	for _, expected := range []string{"package newpkg", "type NewService struct{}", "func (service NewService) Build(worker NewWorker) NewWorker"} {
		if !strings.Contains(service, expected) {
			t.Fatalf("expected %q in moved Go service, got:\n%s", expected, service)
		}
	}
	goConsumer := mustReadFile(t, filepath.Join(root, "internal", "consumer", "api.go"))
	for _, expected := range []string{`"example.com/project/internal/newpkg"`, "func Build() newpkg.NewService", "return newpkg.NewService{}"} {
		if !strings.Contains(goConsumer, expected) {
			t.Fatalf("expected %q in Go consumer, got:\n%s", expected, goConsumer)
		}
	}
	goNoise := mustReadFile(t, filepath.Join(root, "internal", "unrelated", "noise.go"))
	if strings.Contains(goNoise, "newpkg.NewService") {
		t.Fatalf("expected unrelated Go string-like text to remain unchanged, got:\n%s", goNoise)
	}
}
