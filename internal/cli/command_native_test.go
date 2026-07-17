//go:build cgo

package cli

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestMoveUsesNativePHPAnalyzer(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-native-detection"))

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

func TestMoveUsesNativeJavaScriptAnalyzer(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "src", "old-helper.ts"), "export default function helper() {}\n")
	mustWriteFile(t, filepath.Join(root, "src", "consumer.ts"), "import helper from './old-helper';\n")

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath: "src/old-helper.ts",
		NewPath: "src/new-helper.ts",
		DryRun:  true,
		Format:  FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "javascript") {
		t.Fatalf("expected native javascript semantic source, got %#v", report.AutoDetectedAdapters)
	}
	if len(report.Replacements) == 0 {
		t.Fatalf("expected native javascript replacements, got %#v", report)
	}
}

func TestApplyWithNativePHPKeepsImportsBeforeDeclarations(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-import-placement", "before"))

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

	if _, err := os.Stat(filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatch.php")); !os.IsNotExist(err) {
		t.Fatalf("expected original InvoiceBatch path to be moved, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Billing", "Archive", "Domain", "InvoiceBatch.php"), "tests/fixtures/php-import-placement/after/src/Billing/Archive/Domain/InvoiceBatch.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Billing", "Domain", "InvoiceBatchRepository.php"), "tests/fixtures/php-import-placement/after/src/Billing/Domain/InvoiceBatchRepository.php")
}

func TestApplyWithNativePHPUpdatesCaptureMoveImportsAndKeepsFunctionImportGroup(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-capture-move", "before"))

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

	if _, err := os.Stat(filepath.Join(root, "platform", "src", "History", "Capture", "Domain", "Capture.php")); !os.IsNotExist(err) {
		t.Fatalf("expected original Capture path to be moved, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "platform", "src", "History", "Capture.php"), "tests/fixtures/php-capture-move/after/platform/src/History/Capture.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "platform", "src", "History", "Capture", "Domain", "CaptureCollection.php"), "tests/fixtures/php-capture-move/after/platform/src/History/Capture/Domain/CaptureCollection.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "platform", "src", "History", "ComparisonDocument", "Application", "DocumentPageDataMapper.php"), "tests/fixtures/php-capture-move/after/platform/src/History/ComparisonDocument/Application/DocumentPageDataMapper.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "platform", "tests", "History", "ComparisonDocument", "Ui", "Web", "UnifiedPatchRendererTest.php"), "tests/fixtures/php-capture-move/after/platform/tests/History/ComparisonDocument/Ui/Web/UnifiedPatchRendererTest.php")
}

func TestApplyWithNativePHPFailsValidationWhenStaleOldSymbolSurvives(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-stale-symbol"))

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
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-round-trip-imported-types", "before"))

	command := NewCommand()
	report, exitCode := command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Module/Record/Application",
		NewPath:      "src/Module/Record/Mapper",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("first move failed: %d %#v", exitCode, report.Errors)
	}

	report, exitCode = command.runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Module/Record/Mapper",
		NewPath:      "src/Module/Record/Application",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("second move failed: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "Module", "Record", "Mapper", "RecordMapper.php")); !os.IsNotExist(err) {
		t.Fatalf("expected intermediate RecordMapper path to be moved back, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Module", "Record", "Application", "RecordMapper.php"), "tests/fixtures/php-round-trip-imported-types/after/src/Module/Record/Application/RecordMapper.php")
}

func TestApplyWithNativePHPMoveKeepsAliasQualifiedReferencesUnimported(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-alias-qualified", "before"))

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

	if _, err := os.Stat(filepath.Join(root, "src", "Parsing", "SourceDocument.php")); !os.IsNotExist(err) {
		t.Fatalf("expected original SourceDocument path to be moved, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Analysis", "Parsing", "SourceDocument.php"), "tests/fixtures/php-alias-qualified/after/src/Analysis/Parsing/SourceDocument.php")
}

func TestApplyWithNativePHPRepeatedNamespaceMoveDoesNotImportNonClassSegments(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-repeated-namespace-move", "before"))

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/TestStyle",
		NewPath:      "src/PhpSrcTestStyle",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "TestStyle")); !os.IsNotExist(err) {
		t.Fatalf("expected original TestStyle path to be moved, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "PhpSrcTestStyle", "Fixing", "Runner.php"), "tests/fixtures/php-repeated-namespace-move/after/src/PhpSrcTestStyle/Fixing/Runner.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "PhpSrcTestStyle", "Analysis", "Classifier.php"), "tests/fixtures/php-repeated-namespace-move/after/src/PhpSrcTestStyle/Analysis/Classifier.php")
}

func TestApplyWithNativePHPUpdatesImportedConstantsAndFunctions(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-imported-symbols", "before"))

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Config/symbols.php",
		NewPath:      "src/Shared/symbols.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	if _, err := os.Stat(filepath.Join(root, "src", "Config", "symbols.php")); !os.IsNotExist(err) {
		t.Fatalf("expected original symbols path to be moved, got error: %v", err)
	}
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Shared", "symbols.php"), "tests/fixtures/php-imported-symbols/after/src/Shared/symbols.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Http", "Controller.php"), "tests/fixtures/php-imported-symbols/after/src/Http/Controller.php")
}

func TestApplyWithNativePHPUpdatesUnqualifiedConstantsAndFunctions(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-unqualified-symbols", "before"))

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Config/symbols.php",
		NewPath:      "src/Shared/symbols.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Config", "Reader.php"), "tests/fixtures/php-unqualified-symbols/after/src/Config/Reader.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Shared", "symbols.php"), "tests/fixtures/php-unqualified-symbols/after/src/Shared/symbols.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "composer.json"), "tests/fixtures/php-unqualified-symbols/after/composer.json")
}

func TestApplyWithNativePHPWarnsForUnqualifiedConstantsAndFunctionsWithoutComposerFiles(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "php-unqualified-symbols", "no-composer-files", "before"))

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/Config/symbols.php",
		NewPath:      "src/Shared/symbols.php",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Config", "Reader.php"), "tests/fixtures/php-unqualified-symbols/no-composer-files/after/src/Config/Reader.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "src", "Shared", "symbols.php"), "tests/fixtures/php-unqualified-symbols/no-composer-files/after/src/Shared/symbols.php")
	testfixtures.AssertFileMatches(t, filepath.Join(root, "composer.json"), "tests/fixtures/php-unqualified-symbols/no-composer-files/after/composer.json")
	if len(report.Warnings) != 2 {
		t.Fatalf("expected two warnings, got %#v", report.Warnings)
	}
	for _, warning := range report.Warnings {
		if !strings.Contains(warning.Message, "not Composer autoload.files") {
			t.Fatalf("unexpected warning: %#v", warning)
		}
	}
}

func TestApplyWithNativeJavaScriptUpdatesTsConfigAliasProject(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "tsconfig.json"), `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}
`)
	mustWriteFile(t, filepath.Join(root, "src", "old-helper.ts"), `export default function helper() {
  return "ok";
}
`)
	mustWriteFile(t, filepath.Join(root, "src", "consumer.ts"), `import helper from '@/old-helper';

export function run() {
  return helper();
}
`)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/old-helper.ts",
		NewPath:      "src/new-helper.ts",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}

	consumer := mustReadFile(t, filepath.Join(root, "src", "consumer.ts"))
	if !strings.Contains(consumer, "import helper from '@/new-helper';") {
		t.Fatalf("expected alias import rewrite, got:\n%s", consumer)
	}
}

func TestApplyWithNativeJavaScriptUpdatesFixtureProject(t *testing.T) {
	root := copyNamedFixture(t, filepath.Join("tests", "fixtures", "javascript-app"))
	mustWriteFile(t, filepath.Join(root, "vite.config.ts"), `export default {
  resolve: {
    alias: {
      '@ui': `+strconv.Quote(filepath.ToSlash(filepath.Join(root, "src")))+`,
    },
  },
};
`)

	report, exitCode := NewCommand().runWithOptions(t.Context(), root, Options{
		OldPath:      "src/features/checkout/old-card.tsx",
		NewPath:      "src/features/payment/card.tsx",
		Apply:        true,
		NoValidation: true,
		Format:       FormatText,
	}, io.Discard)
	if exitCode != ExitSuccess {
		t.Fatalf("unexpected exit code: %d %#v", exitCode, report.Errors)
	}
	if !hasString(report.AutoDetectedAdapters, "javascript") {
		t.Fatalf("expected native javascript semantic source, got %#v", report.AutoDetectedAdapters)
	}
	if !hasCLIWarning(report.Warnings, "tsconfig.json", `TypeScript path alias "@ambiguous" has multiple targets; skipped conservatively.`) {
		t.Fatalf("expected ambiguous tsconfig warning, got %#v", report.Warnings)
	}
	if !hasCLIWarning(report.Warnings, "package.json", `Package imports entry "#conditional/*" uses conditional targets; skipped conservatively.`) {
		t.Fatalf("expected conditional package warning, got %#v", report.Warnings)
	}
	if _, err := os.Stat(filepath.Join(root, "src", "features", "payment", "card.tsx")); err != nil {
		t.Fatalf("moved javascript file missing: %v", err)
	}

	home := mustReadFile(t, filepath.Join(root, "src", "pages", "home.tsx"))
	for _, expected := range []string{
		"import { CheckoutCard } from '../features/payment/card';",
		"export { CheckoutCard as CardExport } from '../features/payment/card';",
		"const lazy = import('../features/payment/card');",
		"const common = require('../features/payment/card');",
		"import AliasCard from '@app/features/payment/card';",
		"import ViteCard from '@ui/features/payment/card';",
		"import PackageCard from '#app/features/payment/card';",
		"import SelfCard from '@fixture/shop/src/features/payment/card';",
		"import StableTypeScriptAlias from '@checkoutCard';",
		"import StablePackageAlias from '#checkout-card';",
		"import SimilarCard from '../features/checkout/old-card-extra';",
		"import AmbiguousAlias from '@ambiguous';",
		"import ConditionalPackageAlias from '#conditional/features/checkout/old-card';",
		"const dynamic = import(`../features/checkout/${name}`);",
		"const concatenated = require('../features/checkout/' + name);",
	} {
		if !strings.Contains(home, expected) {
			t.Fatalf("expected %q in javascript fixture consumer, got:\n%s", expected, home)
		}
	}

	tsconfig := mustReadFile(t, filepath.Join(root, "tsconfig.json"))
	if !strings.Contains(tsconfig, `"@checkoutCard": ["src/features/payment/card.tsx"]`) {
		t.Fatalf("expected exact tsconfig target rewrite, got:\n%s", tsconfig)
	}
	if !strings.Contains(tsconfig, `"@ambiguous": ["src/features/checkout/old-card.tsx", "src/fallback/card.tsx"]`) {
		t.Fatalf("expected ambiguous tsconfig target to remain unchanged, got:\n%s", tsconfig)
	}

	packageJSON := mustReadFile(t, filepath.Join(root, "package.json"))
	if !strings.Contains(packageJSON, `"#checkout-card": "./src/features/payment/card.tsx"`) {
		t.Fatalf("expected package imports target rewrite, got:\n%s", packageJSON)
	}
	if !strings.Contains(packageJSON, `"#conditional/*": {`) || !strings.Contains(packageJSON, `"default": "./src/*"`) {
		t.Fatalf("expected conditional package imports target to remain unchanged, got:\n%s", packageJSON)
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
