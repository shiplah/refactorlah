package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shiplah/refactorlah/internal/adapters/javascript/rules"
)

type packageSpecifierConfig struct {
	content               []byte
	importTargets         []rules.PackageImportTarget
	conditionalImports    []rules.PackageConditionalImport
	importMappings        []rules.PathAliasMapping
	selfReferenceMappings []rules.PathAliasMapping
}

type rawPackageJSON struct {
	Name    string                     `json:"name"`
	Imports map[string]json.RawMessage `json:"imports"`
}

func readPackageSpecifierConfig(projectRoot string) (packageSpecifierConfig, bool, error) {
	configPath := filepath.Join(projectRoot, "package.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return packageSpecifierConfig{}, false, nil
		}
		return packageSpecifierConfig{}, false, err
	}

	var raw rawPackageJSON
	if err := json.Unmarshal(content, &raw); err != nil {
		return packageSpecifierConfig{}, true, err
	}

	mappings, err := buildPackageImportMappings(projectRoot, raw.Imports)
	if err != nil {
		return packageSpecifierConfig{}, true, err
	}
	return packageSpecifierConfig{
		content:               content,
		importTargets:         buildPackageImportTargets(raw.Imports),
		conditionalImports:    buildPackageConditionalImports(raw.Imports),
		importMappings:        mappings,
		selfReferenceMappings: rules.PackageSelfReferenceMappings(raw.Name),
	}, true, nil
}
