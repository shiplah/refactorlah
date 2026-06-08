package staticimports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
)

func TestScannerRewritesStaticImportSpecifiers(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "assets/app.js", `import '../src/Old/style.css';
export { thing } from "../src/Old/style.css";
const lazy = import('../src/Old/style.css');
const style = require("../src/Old/style.css");
`)

	replacements, err := Scanner{}.Scan(root, []string{"assets/app.js"}, []planning.FileMove{{
		OldPath: "src/Old/style.css",
		NewPath: "src/New/style.css",
	}})
	if err != nil {
		t.Fatalf("scan static imports: %v", err)
	}
	if len(replacements) != 4 {
		t.Fatalf("expected four replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if replacement.Replacement != "../src/New/style.css" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestScannerRewritesCssImports(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "assets/app.css", `@import "../src/Old/style.css";`)

	replacements, err := Scanner{}.Scan(root, []string{"assets/app.css"}, []planning.FileMove{{
		OldPath: "src/Old/style.css",
		NewPath: "src/New/style.css",
	}})
	if err != nil {
		t.Fatalf("scan static imports: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
}

func TestScannerSupportsBareSameDirectorySpecifiers(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "assets/app.js", `import 'old.css';`)

	replacements, err := Scanner{}.Scan(root, []string{"assets/app.js"}, []planning.FileMove{{
		OldPath: "assets/old.css",
		NewPath: "assets/new.css",
	}})
	if err != nil {
		t.Fatalf("scan static imports: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "new.css" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestScannerScanModulesRewritesExtensionlessImportAndRequireSpecifiers(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "src/consumer.ts", `import helper from './old-helper';
const loaded = require("./old-helper");
`)

	replacements, err := Scanner{}.ScanModules(root, []string{"src/consumer.ts"}, []planning.FileMove{{
		OldPath: "src/old-helper.ts",
		NewPath: "src/new-helper.ts",
	}})
	if err != nil {
		t.Fatalf("scan module imports: %v", err)
	}
	if len(replacements) != 2 {
		t.Fatalf("expected two replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if replacement.Replacement != "./new-helper" {
			t.Fatalf("unexpected replacement %q", replacement.Replacement)
		}
	}
}

func TestScannerScanModulesRewritesDirectoryImportForMovedIndexModule(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "src/consumer.ts", `import helper from './old';
`)

	replacements, err := Scanner{}.ScanModules(root, []string{"src/consumer.ts"}, []planning.FileMove{{
		OldPath: "src/old/index.ts",
		NewPath: "src/new/index.ts",
	}})
	if err != nil {
		t.Fatalf("scan module imports: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "./new" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestScannerSkipsDynamicImports(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "assets/app.js", `import('../src/' + name + '.css');`)

	replacements, err := Scanner{}.Scan(root, []string{"assets/app.js"}, []planning.FileMove{{
		OldPath: "src/Old/style.css",
		NewPath: "src/New/style.css",
	}})
	if err != nil {
		t.Fatalf("scan static imports: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestCandidateNeedlesIncludesOldPathAndBasename(t *testing.T) {
	needles := CandidateNeedles([]planning.FileMove{{
		OldPath: "src/Old/style.css",
		NewPath: "src/New/style.css",
	}})

	for _, expected := range []string{"src/Old/style.css", "style.css"} {
		if !containsStaticNeedle(needles, expected) {
			t.Fatalf("expected needle %q in %#v", expected, needles)
		}
	}
}

func TestModuleCandidateNeedlesIncludeExtensionlessModuleNames(t *testing.T) {
	needles := ModuleCandidateNeedles([]planning.FileMove{{
		OldPath: "src/old/index.ts",
		NewPath: "src/new/index.ts",
	}})

	for _, expected := range []string{"src/old/index.ts", "index.ts", "src/old", "old"} {
		if !containsStaticNeedle(needles, expected) {
			t.Fatalf("expected module needle %q in %#v", expected, needles)
		}
	}
}

func containsStaticNeedle(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func writeStaticImportFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
