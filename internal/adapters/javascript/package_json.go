package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

const (
	packageImportsReason = "javascript-package-imports"
	packageImportsRule   = "javascript.PackageImportsRule"
)

type packageImportsConfig struct {
	mappings []pathAliasMapping
}

type rawPackageJSON struct {
	Imports map[string]json.RawMessage `json:"imports"`
}

func readPackageImportsConfig(projectRoot string) (packageImportsConfig, bool, error) {
	configPath := filepath.Join(projectRoot, "package.json")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return packageImportsConfig{}, false, nil
		}
		return packageImportsConfig{}, false, err
	}

	var raw rawPackageJSON
	if err := json.Unmarshal(content, &raw); err != nil {
		return packageImportsConfig{}, true, err
	}

	mappings, err := buildPackageImportMappings(projectRoot, raw.Imports)
	if err != nil {
		return packageImportsConfig{}, true, err
	}
	return packageImportsConfig{mappings: mappings}, true, nil
}

func buildPackageImportMappings(projectRoot string, imports map[string]json.RawMessage) ([]pathAliasMapping, error) {
	if len(imports) == 0 {
		return nil, nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var mappings []pathAliasMapping
	for _, importKey := range importKeys {
		if !strings.HasPrefix(importKey, "#") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err != nil {
			continue
		}
		if !strings.HasPrefix(target, "./") {
			continue
		}

		aliasPrefix, aliasOK := wildcardPrefix(importKey)
		targetPrefix, targetOK := wildcardPrefix(target)
		if !aliasOK || !targetOK {
			continue
		}

		resolvedPrefix, ok, err := resolveAliasTargetPrefix(projectRoot, projectRoot, targetPrefix)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		mappings = append(mappings, pathAliasMapping{
			aliasPrefix:  aliasPrefix,
			targetPrefix: resolvedPrefix,
		})
	}

	return mappings, nil
}

func packageImportsSpecifierRewrites(config packageImportsConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return specifierRewritesForPathAliases(config.mappings, moves, packageImportsReason, packageImportsRule)
}
