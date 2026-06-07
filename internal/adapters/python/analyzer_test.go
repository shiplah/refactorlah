//go:build cgo

package python

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func TestAnalyzerUpdatesAbsoluteAndRelativeImports(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "src/collector/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/snapshot_manifest.py", "class SnapshotManifest: pass\n")
	writePythonFixture(t, root, "pyproject.toml", `handler = "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest"`)
	writePythonFixture(t, root, "src/collector/assembly/cache_files/loader.py", `from collector.assembly.cache_files.snapshot_manifest import SnapshotManifest
from collector.assembly.cache_files import snapshot_manifest
from .snapshot_manifest import SnapshotManifest as LocalSnapshotManifest
from . import snapshot_manifest as manifest

manifest_module = snapshot_manifest.load()
qualified_manifest = collector.assembly.cache_files.snapshot_manifest.load()

def typed_manifest() -> "collector.assembly.cache_files.snapshot_manifest.SnapshotManifest":
    return manifest_module
`)

	response, relevant, err := analyzePython(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/collector/assembly/cache_files/snapshot_manifest.py",
			NewPath: "src/collector/assembly/cache_files/summary_manifest.py",
		}},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}
	if !relevant {
		t.Fatal("expected python analyzer to be relevant")
	}
	if len(response.SymbolMappings) != 1 {
		t.Fatalf("expected 1 symbol mapping, got %#v", response.SymbolMappings)
	}

	assertPythonReplacement(t, response.Replacements, "src/collector/assembly/cache_files/loader.py", "collector.assembly.cache_files.summary_manifest")
	assertPythonReplacement(t, response.Replacements, "src/collector/assembly/cache_files/loader.py", "summary_manifest")
	assertPythonReplacement(t, response.Replacements, "pyproject.toml", "collector.assembly.cache_files.summary_manifest")
}

func TestAnalyzerWarnsForPythonFileOutsideSourceRoots(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "tools/snapshot_manifest.py", "class SnapshotManifest: pass\n")

	response, relevant, err := analyzePython(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "tools/snapshot_manifest.py",
			NewPath: "other/summary_manifest.py",
		}},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}
	if !relevant {
		t.Fatal("expected python analyzer to be relevant")
	}
	if len(response.Warnings) != 0 {
		t.Fatalf("fallback source root should derive modules without warnings, got %#v", response.Warnings)
	}
}

func TestAnalyzerHonoursScanExcludes(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "src/app/services/billing.py", "class InvoiceService: pass\n")
	writePythonFixture(t, root, "src/app/http/controller.py", "import app.services.billing\n")
	writePythonFixture(t, root, "src/app/generated/fixture.py", "import app.services.billing\n")

	response, _, err := analyzePythonWithConfig(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/app/services/billing.py",
			NewPath: "src/app/domain/invoicing.py",
		}},
	}, config.Config{
		Exclude: []string{"src/app/generated/**"},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}

	assertPythonReplacement(t, response.Replacements, "src/app/http/controller.py", "app.domain.invoicing")
	assertNoPythonReplacementInFile(t, response.Replacements, "src/app/generated/fixture.py")
}

func TestAnalyzerUpdatesFixtureProject(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "src/app/__init__.py", "")
	writePythonFixture(t, root, "src/app/services/__init__.py", "")
	writePythonFixture(t, root, "src/app/services/billing.py", "class InvoiceService:\n    pass\n")
	writePythonFixture(t, root, "src/app/http/controller.py", `import importlib

import app.services.billing
from app.services import billing as billing_module
from app.services.billing import InvoiceService


def build() -> "app.services.billing.InvoiceService":
    service = app.services.billing.InvoiceService()
    alias_service = billing_module.InvoiceService()
    imported_service = InvoiceService()
    literal = "app.services.billing.InvoiceService"
    config = {"handler": "app.services.billing.InvoiceService"}
    importlib.import_module("dynamic.name")
    return service or alias_service or imported_service or literal or config
`)
	writePythonFixture(t, root, "src/app/services/consumer.py", `from . import billing
from .billing import InvoiceService


def build_relative() -> InvoiceService:
    return billing.InvoiceService()
`)
	writePythonFixture(t, root, "src/app/generated/fixture.py", `import app.services.billing


def generated() -> app.services.billing.InvoiceService:
    return app.services.billing.InvoiceService()
`)
	writePythonFixture(t, root, "pyproject.toml", "[tool.example]\nhandler = \"app.services.billing.InvoiceService\"\n")
	writePythonFixture(t, root, "config/routes.yaml", "billing_handler: app.services.billing.InvoiceService\n# app.services.billing.CommentOnly\n")

	response, _, err := analyzePythonWithConfig(t, root, planning.MovePlan{
		Moves: []planning.FileMove{{
			OldPath: "src/app/services/billing.py",
			NewPath: "src/app/domain/invoicing.py",
		}},
	}, config.Config{
		Exclude: []string{"src/app/generated/**"},
	})
	if err != nil {
		t.Fatalf("analyze python: %v", err)
	}

	controller := mustReadPythonFixture(t, root, "src/app/http/controller.py")
	updatedController := applyAdapterReplacements(controller, response.Replacements, "src/app/http/controller.py")
	if !strings.Contains(updatedController, "import app.domain.invoicing") {
		t.Fatalf("expected absolute import rewrite, got:\n%s", updatedController)
	}
	if !strings.Contains(updatedController, "from app.domain import invoicing as billing_module") {
		t.Fatalf("expected aliased parent import rewrite, got:\n%s", updatedController)
	}
	if !strings.Contains(updatedController, `def build() -> "app.domain.invoicing.InvoiceService":`) {
		t.Fatalf("expected string annotation rewrite, got:\n%s", updatedController)
	}
	if !strings.Contains(updatedController, `literal = "app.services.billing.InvoiceService"`) {
		t.Fatalf("expected arbitrary string to remain unchanged, got:\n%s", updatedController)
	}
	if !strings.Contains(updatedController, `config = {"handler": "app.services.billing.InvoiceService"}`) {
		t.Fatalf("expected dictionary value string to remain unchanged, got:\n%s", updatedController)
	}

	relativeConsumer := applyAdapterReplacements(
		mustReadPythonFixture(t, root, "src/app/services/consumer.py"),
		response.Replacements,
		"src/app/services/consumer.py",
	)
	if !strings.Contains(relativeConsumer, "from app.domain import invoicing") {
		t.Fatalf("expected relative import rewrite, got:\n%s", relativeConsumer)
	}
	if !strings.Contains(relativeConsumer, "return invoicing.InvoiceService()") {
		t.Fatalf("expected imported module reference rewrite, got:\n%s", relativeConsumer)
	}

	assertNoPythonReplacementInFile(t, response.Replacements, "src/app/generated/fixture.py")
	assertPythonUpdatedTextContains(t, root, response.Replacements, "pyproject.toml", `handler = "app.domain.invoicing.InvoiceService"`)
	assertPythonUpdatedTextContains(t, root, response.Replacements, "config/routes.yaml", "billing_handler: app.domain.invoicing.InvoiceService")
	assertPythonUpdatedTextContains(t, root, response.Replacements, "config/routes.yaml", "# app.services.billing.CommentOnly")
	assertPythonWarning(t, response.Warnings, "src/app/http/controller.py", "Dynamic Python import detected; not changed.")
}

func assertPythonReplacement(t *testing.T, replacements []adapterproto.Replacement, file string, newText string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file && replacement.Replacement == newText {
			return
		}
	}
	t.Fatalf("expected replacement in %s to %q, got %#v", file, newText, replacements)
}

func assertNoPythonReplacementInFile(t *testing.T, replacements []adapterproto.Replacement, file string) {
	t.Helper()

	for _, replacement := range replacements {
		if replacement.File == file {
			t.Fatalf("unexpected replacement in excluded file %s: %#v", file, replacement)
		}
	}
}

func assertPythonUpdatedTextContains(t *testing.T, root string, replacements []adapterproto.Replacement, file string, expected string) {
	t.Helper()

	updated := applyAdapterReplacements(mustReadPythonFixture(t, root, file), replacements, file)
	if !strings.Contains(updated, expected) {
		t.Fatalf("expected %s to contain %q, got:\n%s", file, expected, updated)
	}
}

func assertPythonWarning(t *testing.T, warnings []adapterproto.Warning, file string, message string) {
	t.Helper()

	for _, warning := range warnings {
		if warning.File == file && warning.Message == message {
			return
		}
	}
	t.Fatalf("expected warning in %s: %s, got %#v", file, message, warnings)
}

func mustReadPythonFixture(t *testing.T, root string, relativePath string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func applyAdapterReplacements(content string, replacements []adapterproto.Replacement, file string) string {
	fileReplacements := make([]adapterproto.Replacement, 0, len(replacements))
	for _, replacement := range replacements {
		if replacement.File == file {
			fileReplacements = append(fileReplacements, replacement)
		}
	}
	sort.Slice(fileReplacements, func(left int, right int) bool {
		return fileReplacements[left].Start > fileReplacements[right].Start
	})

	result := []byte(content)
	for _, replacement := range fileReplacements {
		next := make([]byte, 0, len(result)-replacement.End+replacement.Start+len(replacement.Replacement))
		next = append(next, result[:replacement.Start]...)
		next = append(next, []byte(replacement.Replacement)...)
		next = append(next, result[replacement.End:]...)
		result = next
	}
	return string(result)
}
