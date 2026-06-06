package twig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigReaderReadsYamlTwigRoots(t *testing.T) {
	root := t.TempDir()
	writeTwigFixture(t, root, "config/packages/twig.yaml", `twig:
  default_path: '%kernel.project_dir%/templates'
  paths:
    '%kernel.project_dir%/src/Billing/Archive/Listing/Ui/Web/Twig': Billing
`)

	configuration, err := ConfigReader{}.Read(root)
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "templates"})
	assertTwigRoot(t, configuration, PathRoot{Path: "src/Billing/Archive/Listing/Ui/Web/Twig", Namespace: "Billing"})
}

func TestConfigReaderReadsPhpTwigRoots(t *testing.T) {
	root := t.TempDir()
	writeTwigFixture(t, root, "config/packages/twig.php", `<?php
$twig->defaultPath('%kernel.project_dir%/templates');
$twig->path('%kernel.project_dir%/src/Shared/Ui/Web/Twig', 'Shared');
`)

	configuration, err := ConfigReader{}.Read(root)
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "templates"})
	assertTwigRoot(t, configuration, PathRoot{Path: "src/Shared/Ui/Web/Twig", Namespace: "Shared"})
}

func TestConfigReaderFallsBackToTemplatesDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "templates"), 0o755); err != nil {
		t.Fatal(err)
	}

	configuration, err := ConfigReader{}.Read(root)
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "templates"})
}

func TestConfigReaderPrefixesRootsFromComposerSubdirectory(t *testing.T) {
	root := t.TempDir()
	writeTwigFixture(t, root, "platform/config/packages/twig.yaml", `twig:
  default_path: '%kernel.project_dir%/templates'
`)

	configuration, err := ConfigReader{}.ReadFromConfigRoot(root, filepath.Join(root, "platform"))
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "platform/templates"})
}

func writeTwigFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertTwigRoot(t *testing.T, configuration PathConfiguration, expected PathRoot) {
	t.Helper()

	for _, root := range configuration.Roots {
		if root == expected {
			return
		}
	}

	t.Fatalf("missing twig root %#v in %#v", expected, configuration.Roots)
}
