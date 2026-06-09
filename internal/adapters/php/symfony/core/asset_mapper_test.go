package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
)

func TestProjectDirectoryPathMappingsDerivesDirectoryMapping(t *testing.T) {
	mappings := ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Shared/Ui/Web",
		NewPath: "src/Shared/Ui/Browser",
		IsDir:   true,
	})

	if len(mappings) != 1 {
		t.Fatalf("expected one mapping, got %#v", mappings)
	}
	if mappings[0].OldReference != "src/Shared/Ui/Web/" || mappings[0].NewReference != "src/Shared/Ui/Browser/" {
		t.Fatalf("unexpected mapping %#v", mappings[0])
	}
}

func TestProjectDirectoryPathMappingsSkipsFileMoves(t *testing.T) {
	mappings := ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Shared/Ui/Web/file.css",
		NewPath: "src/Shared/Ui/Browser/file.css",
		IsDir:   false,
	})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
}

func TestAssetMapperScannerRewritesExactYamlAssetMapperPaths(t *testing.T) {
	root := t.TempDir()
	writeAssetMapperFixture(t, root, "config/packages/asset_mapper.yaml", `framework:
  asset_mapper:
    paths:
      - 'src/Shared/Ui/Web/'
`)

	replacements, err := AssetMapperScanner{}.Scan(root, []string{"config/packages/asset_mapper.yaml"}, ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Shared/Ui/Web",
		NewPath: "src/Shared/Ui/Browser",
		IsDir:   true,
	}))
	if err != nil {
		t.Fatalf("scan asset mapper: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "'src/Shared/Ui/Browser/'" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestAssetMapperScannerSkipsNonAssetMapperYaml(t *testing.T) {
	root := t.TempDir()
	writeAssetMapperFixture(t, root, "config/packages/example.yaml", `paths:
  - 'src/Shared/Ui/Web/'
`)

	replacements, err := AssetMapperScanner{}.Scan(root, []string{"config/packages/example.yaml"}, ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Shared/Ui/Web",
		NewPath: "src/Shared/Ui/Browser",
		IsDir:   true,
	}))
	if err != nil {
		t.Fatalf("scan asset mapper: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func writeAssetMapperFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()

	absolutePath := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
