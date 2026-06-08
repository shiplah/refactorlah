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
	packageImportsReason       = "javascript-package-imports"
	packageImportsRule         = "javascript.PackageImportsRule"
	packageSelfReferenceReason = "javascript-package-self-reference"
	packageSelfReferenceRule   = "javascript.PackageSelfReferenceRule"
)

type packageSpecifierConfig struct {
	importMappings        []pathAliasMapping
	selfReferenceMappings []pathAliasMapping
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
		importMappings:        mappings,
		selfReferenceMappings: packageSelfReferenceMappings(raw.Name),
	}, true, nil
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

func packageSelfReferenceMappings(packageName string) []pathAliasMapping {
	aliasPrefix, ok := packageSelfReferenceAliasPrefix(packageName)
	if !ok {
		return nil
	}
	return []pathAliasMapping{{
		aliasPrefix: aliasPrefix,
	}}
}

func packageSelfReferenceAliasPrefix(packageName string) (string, bool) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" || strings.HasPrefix(packageName, ".") || strings.HasPrefix(packageName, "/") {
		return "", false
	}
	if strings.ContainsAny(packageName, "\\ \t\r\n") {
		return "", false
	}

	parts := strings.Split(packageName, "/")
	if strings.HasPrefix(packageName, "@") {
		if len(parts) != 2 || parts[0] == "@" || parts[1] == "" {
			return "", false
		}
		return packageName + "/", true
	}
	if len(parts) != 1 {
		return "", false
	}
	return packageName + "/", true
}

func packageSpecifierRewrites(config packageSpecifierConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := specifierRewritesForPathAliases(config.importMappings, moves, packageImportsReason, packageImportsRule)
	rewrites = append(rewrites, specifierRewritesForPathAliases(config.selfReferenceMappings, moves, packageSelfReferenceReason, packageSelfReferenceRule)...)
	return rewrites
}
