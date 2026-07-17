package core

import (
	"testing"

	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/testfixtures"
)

func TestProjectDirectoryPathMappingsDerivesDirectoryMapping(t *testing.T) {
	mappings := ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Module/Ui/Web",
		NewPath: "src/Module/Ui/Browser",
		IsDir:   true,
	})

	if len(mappings) != 1 {
		t.Fatalf("expected one mapping, got %#v", mappings)
	}
	if mappings[0].OldReference != "src/Module/Ui/Web/" || mappings[0].NewReference != "src/Module/Ui/Browser/" {
		t.Fatalf("unexpected mapping %#v", mappings[0])
	}
}

func TestProjectDirectoryPathMappingsSkipsFileMoves(t *testing.T) {
	mappings := ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Module/Ui/Web/file.css",
		NewPath: "src/Module/Ui/Browser/file.css",
		IsDir:   false,
	})

	if len(mappings) != 0 {
		t.Fatalf("expected no mappings, got %#v", mappings)
	}
}

func TestAssetMapperScannerRewritesExactYamlAssetMapperPaths(t *testing.T) {
	root := assetMapperFixtureRoot(t, "exact")

	replacements, err := AssetMapperScanner{}.Scan(root, []string{"config/packages/asset_mapper.yaml"}, ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Module/Ui/Web",
		NewPath: "src/Module/Ui/Browser",
		IsDir:   true,
	}))
	if err != nil {
		t.Fatalf("scan asset mapper: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected one replacement, got %#v", replacements)
	}
	if replacements[0].Replacement != "'src/Module/Ui/Browser/'" {
		t.Fatalf("unexpected replacement %q", replacements[0].Replacement)
	}
}

func TestAssetMapperScannerSkipsNonAssetMapperYaml(t *testing.T) {
	root := assetMapperFixtureRoot(t, "non-asset-mapper")

	replacements, err := AssetMapperScanner{}.Scan(root, []string{"config/packages/example.yaml"}, ProjectDirectoryPathMappings(planning.MovePlan{
		OldPath: "src/Module/Ui/Web",
		NewPath: "src/Module/Ui/Browser",
		IsDir:   true,
	}))
	if err != nil {
		t.Fatalf("scan asset mapper: %v", err)
	}
	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func assetMapperFixtureRoot(t *testing.T, scenario string) string {
	t.Helper()

	return testfixtures.CopyDir(t, "tests/fixtures/php-symfony-core/asset-mapper/"+scenario)
}
