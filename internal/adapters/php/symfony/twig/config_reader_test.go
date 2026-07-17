package twig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestConfigReaderReadsYamlTwigRoots(t *testing.T) {
	root := twigFixtureRoot(t, "config-reader/yaml")

	configuration, err := ConfigReader{}.Read(root)
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "templates"})
	assertTwigRoot(t, configuration, PathRoot{Path: "src/Module/List/Ui/Twig", Namespace: "Module"})
}

func TestConfigReaderReadsPhpTwigRoots(t *testing.T) {
	root := twigFixtureRoot(t, "config-reader/php")

	configuration, err := ConfigReader{}.Read(root)
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "templates"})
	assertTwigRoot(t, configuration, PathRoot{Path: "src/Common/Ui/Twig", Namespace: "Common"})
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
	root := twigFixtureRoot(t, "config-reader/composer-subdirectory")

	configuration, err := ConfigReader{}.ReadFromConfigRoot(root, filepath.Join(root, "platform"))
	if err != nil {
		t.Fatalf("read twig config: %v", err)
	}

	assertTwigRoot(t, configuration, PathRoot{Path: "platform/templates"})
}

func twigFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-symfony-twig/"+scenario)
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
