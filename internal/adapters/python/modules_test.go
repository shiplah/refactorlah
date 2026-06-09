package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
)

func TestSourceRootResolverDetectsSrcAndPackageParents(t *testing.T) {
	root := t.TempDir()
	writePythonFixture(t, root, "src/collector/assembly/__init__.py", "")
	writePythonFixture(t, root, "src/collector/assembly/cache_files/__init__.py", "")

	roots, err := SourceRootResolver{}.Resolve(root, []planning.FileMove{{
		OldPath: "src/collector/assembly/cache_files/snapshot_manifest.py",
		NewPath: "src/collector/assembly/cache_files/summary_manifest.py",
	}})
	if err != nil {
		t.Fatalf("resolve source roots: %v", err)
	}

	if !containsString(roots, "src") {
		t.Fatalf("expected src source root, got %#v", roots)
	}
}

func TestModuleMapperDerivesModuleMappings(t *testing.T) {
	mapper := NewModuleMapper([]string{"src"})

	mappings, warnings := mapper.Derive([]planning.FileMove{{
		OldPath: "src/collector/assembly/cache_files/snapshot_manifest.py",
		NewPath: "src/collector/assembly/cache_files/summary_manifest.py",
	}})

	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %#v", mappings)
	}
	if mappings[0].OldModule != "collector.assembly.cache_files.snapshot_manifest" {
		t.Fatalf("unexpected old module %q", mappings[0].OldModule)
	}
	if mappings[0].NewModule != "collector.assembly.cache_files.summary_manifest" {
		t.Fatalf("unexpected new module %q", mappings[0].NewModule)
	}
}

func TestModuleMapperHandlesPackageInitModules(t *testing.T) {
	mapper := NewModuleMapper([]string{"src"})

	module, ok := mapper.ModuleForPath("src/collector/assembly/__init__.py")
	if !ok {
		t.Fatal("expected module")
	}
	if module != "collector.assembly" {
		t.Fatalf("unexpected module %q", module)
	}

	packageName, ok := mapper.PackageForPath("src/collector/assembly/__init__.py")
	if !ok {
		t.Fatal("expected package")
	}
	if packageName != "collector.assembly" {
		t.Fatalf("unexpected package %q", packageName)
	}
}

func writePythonFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
