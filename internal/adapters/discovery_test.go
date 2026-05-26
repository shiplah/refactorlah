package adapters

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoveryFindsProjectLocalPHPAdapter(t *testing.T) {
	root := t.TempDir()
	adapterPath := filepath.Join(root, "adapters", "php", "bin", "refactorlah-php")
	if err := os.MkdirAll(filepath.Dir(adapterPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adapterPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	discovery := NewDiscovery()
	foundPath, ok := discovery.FindPHPAdapter(root)
	if !ok {
		t.Fatal("expected adapter to be found")
	}
	if foundPath != adapterPath {
		t.Fatalf("expected %s, got %s", adapterPath, foundPath)
	}
}

func TestDiscoveryFindsProjectLocalPythonAdapter(t *testing.T) {
	root := t.TempDir()
	adapterPath := filepath.Join(root, "adapters", "python", "bin", "refactorlah-python")
	if err := os.MkdirAll(filepath.Dir(adapterPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(adapterPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	discovery := NewDiscovery()
	foundPath, ok := discovery.FindPythonAdapter(root)
	if !ok {
		t.Fatal("expected adapter to be found")
	}
	if foundPath != adapterPath {
		t.Fatalf("expected %s, got %s", adapterPath, foundPath)
	}
}

func TestRequirePHPAdapterFailsBeforeExecutionWhenRuntimeIsMissing(t *testing.T) {
	root := t.TempDir()
	writePHPAdapter(t, root, false)

	discovery := &Discovery{
		lookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
		run: func(context.Context, string, ...string) error {
			t.Fatal("runtime check should not run when php is missing")
			return nil
		},
	}

	_, err := discovery.RequirePHPAdapter(t.Context(), root)
	if err == nil {
		t.Fatal("expected missing php runtime error")
	}
	if !errors.Is(err, ErrAdapterFailure) {
		t.Fatalf("expected adapter failure, got %v", err)
	}
	if !strings.Contains(err.Error(), "php was not found") {
		t.Fatalf("expected missing php guidance, got %v", err)
	}
}

func TestRequirePHPAdapterFailsWhenDependenciesAreMissing(t *testing.T) {
	root := t.TempDir()
	writePHPAdapter(t, root, false)

	discovery := &Discovery{
		lookPath: func(string) (string, error) {
			return "/usr/bin/php", nil
		},
		run: func(context.Context, string, ...string) error {
			return nil
		},
	}

	_, err := discovery.RequirePHPAdapter(t.Context(), root)
	if err == nil {
		t.Fatal("expected missing dependencies error")
	}
	if !strings.Contains(err.Error(), "dependency is missing") {
		t.Fatalf("expected dependency guidance, got %v", err)
	}
}

func TestRequirePHPAdapterPassesWhenRuntimeAndDependenciesAreReady(t *testing.T) {
	root := t.TempDir()
	adapterPath := writePHPAdapter(t, root, true)

	discovery := &Discovery{
		lookPath: func(string) (string, error) {
			return "/usr/bin/php", nil
		},
		run: func(context.Context, string, ...string) error {
			return nil
		},
	}

	foundPath, err := discovery.RequirePHPAdapter(t.Context(), root)
	if err != nil {
		t.Fatalf("expected ready adapter, got %v", err)
	}
	if foundPath != adapterPath {
		t.Fatalf("expected %s, got %s", adapterPath, foundPath)
	}
}

func TestRequirePythonAdapterFailsBeforeExecutionWhenRuntimeIsMissing(t *testing.T) {
	root := t.TempDir()
	writePythonAdapter(t, root, false)

	discovery := &Discovery{
		lookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
		run: func(context.Context, string, ...string) error {
			t.Fatal("runtime check should not run when python3 is missing")
			return nil
		},
	}

	_, err := discovery.RequirePythonAdapter(t.Context(), root)
	if err == nil {
		t.Fatal("expected missing python runtime error")
	}
	if !errors.Is(err, ErrAdapterFailure) {
		t.Fatalf("expected adapter failure, got %v", err)
	}
	if !strings.Contains(err.Error(), "python3 was not found") {
		t.Fatalf("expected missing python guidance, got %v", err)
	}
}

func TestRequirePythonAdapterPassesWhenRuntimeAndFilesAreReady(t *testing.T) {
	root := t.TempDir()
	adapterPath := writePythonAdapter(t, root, true)

	discovery := &Discovery{
		lookPath: func(string) (string, error) {
			return "/usr/bin/python3", nil
		},
		run: func(context.Context, string, ...string) error {
			return nil
		},
	}

	foundPath, err := discovery.RequirePythonAdapter(t.Context(), root)
	if err != nil {
		t.Fatalf("expected ready adapter, got %v", err)
	}
	if foundPath != adapterPath {
		t.Fatalf("expected %s, got %s", adapterPath, foundPath)
	}
}

func TestLoadManifestRejectsIncompleteAdapterManifest(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "adapter.json"), `{"name":"broken"}`)

	_, err := LoadManifest(root)
	if err == nil {
		t.Fatal("expected incomplete manifest error")
	}
	if !strings.Contains(err.Error(), "manifest is incomplete") {
		t.Fatalf("expected incomplete manifest guidance, got %v", err)
	}
}

func writePHPAdapter(t *testing.T, root string, withDependencies bool) string {
	t.Helper()
	adapterRoot := filepath.Join(root, "adapters", "php")
	adapterPath := filepath.Join(adapterRoot, "bin", "refactorlah-php")
	writeExecutable(t, adapterPath)
	writeFile(t, filepath.Join(adapterRoot, "adapter.json"), `{
  "name": "php",
  "executable": "refactorlah-php",
  "runtime": {
    "command": "php",
    "minimumVersion": "8.2.0",
    "versionCheck": ["-r", "exit(version_compare(PHP_VERSION, '8.2.0', '>=') ? 0 : 1);"]
  },
  "requiredFiles": ["vendor/autoload.php"]
}`)
	if withDependencies {
		writeFile(t, filepath.Join(adapterRoot, "vendor", "autoload.php"), "<?php\n")
	}
	return adapterPath
}

func writePythonAdapter(t *testing.T, root string, withFiles bool) string {
	t.Helper()
	adapterRoot := filepath.Join(root, "adapters", "python")
	adapterPath := filepath.Join(adapterRoot, "bin", "refactorlah-python")
	writeExecutable(t, adapterPath)
	writeFile(t, filepath.Join(adapterRoot, "adapter.json"), `{
  "name": "python",
  "executable": "refactorlah-python",
  "runtime": {
    "command": "python3",
    "minimumVersion": "3.11",
    "versionCheck": ["-c", "import sys; raise SystemExit(0 if sys.version_info >= (3, 11) else 1)"]
  },
  "requiredFiles": ["src/analyze_command.py"]
}`)
	if withFiles {
		writeFile(t, filepath.Join(adapterRoot, "src", "analyze_command.py"), "")
	}
	return adapterPath
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	writeFile(t, path, "#!/bin/sh\n")
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
