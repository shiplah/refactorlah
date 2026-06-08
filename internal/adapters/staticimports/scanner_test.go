package staticimports

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NickSdot/refactorlah/internal/planning"
)

func TestScannerRewritesStaticImportSpecifiers(t *testing.T) {
	root := t.TempDir()
	writeStaticImportFixture(t, root, "assets/app.js", `import '../src/Old/style.css';
export { thing } from "../src/Old/style.css";
const lazy = import('../src/Old/style.css');
`)

	replacements, err := Scanner{}.Scan(root, []string{"assets/app.js"}, []planning.FileMove{{
		OldPath: "src/Old/style.css",
		NewPath: "src/New/style.css",
	}})
	if err != nil {
		t.Fatalf("scan static imports: %v", err)
	}
	if len(replacements) != 3 {
		t.Fatalf("expected three replacements, got %#v", replacements)
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
