package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
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

func buildPackageImportMappings(projectRoot string, imports map[string]json.RawMessage) ([]rules.PathAliasMapping, error) {
	if len(imports) == 0 {
		return nil, nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var mappings []rules.PathAliasMapping
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

		aliasPrefix, aliasOK := rules.WildcardPrefix(importKey)
		targetPrefix, targetOK := rules.WildcardPrefix(target)
		if !aliasOK || !targetOK {
			continue
		}

		resolvedPrefix, ok, err := rules.ResolveAliasTargetPrefix(projectRoot, projectRoot, targetPrefix)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		mappings = append(mappings, rules.PathAliasMapping{
			AliasPrefix:  aliasPrefix,
			TargetPrefix: resolvedPrefix,
		})
	}

	return mappings, nil
}

func buildPackageImportTargets(imports map[string]json.RawMessage) []rules.PackageImportTarget {
	if len(imports) == 0 {
		return nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var targets []rules.PackageImportTarget
	for _, importKey := range importKeys {
		if !strings.HasPrefix(importKey, "#") || strings.Contains(importKey, "*") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err != nil {
			continue
		}
		if !strings.HasPrefix(target, "./") || strings.Contains(target, "*") {
			continue
		}
		if !rules.IsJavaScriptModuleExtension(filepath.Ext(target)) {
			continue
		}

		targets = append(targets, rules.PackageImportTarget{Target: target})
	}
	return targets
}

func buildPackageConditionalImports(imports map[string]json.RawMessage) []rules.PackageConditionalImport {
	if len(imports) == 0 {
		return nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var conditionalImports []rules.PackageConditionalImport
	for _, importKey := range importKeys {
		if !strings.HasPrefix(importKey, "#") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err == nil {
			continue
		}

		targets := packageTargetStrings(imports[importKey])
		if len(targets) == 0 {
			continue
		}
		conditionalImports = append(conditionalImports, rules.PackageConditionalImport{
			Key:     importKey,
			Targets: targets,
		})
	}
	return conditionalImports
}

func packageTargetStrings(raw json.RawMessage) []string {
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var targets []string
	collectPackageTargetStrings(value, seen, &targets)
	sort.Strings(targets)
	return targets
}

func collectPackageTargetStrings(value interface{}, seen map[string]bool, targets *[]string) {
	switch typed := value.(type) {
	case string:
		if !strings.HasPrefix(typed, "./") || seen[typed] {
			return
		}
		seen[typed] = true
		*targets = append(*targets, typed)
	case []interface{}:
		for _, item := range typed {
			collectPackageTargetStrings(item, seen, targets)
		}
	case map[string]interface{}:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectPackageTargetStrings(typed[key], seen, targets)
		}
	}
}

func packageSpecifierRewrites(config packageSpecifierConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := rules.PackageImportAliasRule{}.Collect(config.importMappings, config.selfReferenceMappings, moves)
	return rewrites
}

func packageImportTargetReplacements(config packageSpecifierConfig, moves []planning.FileMove) []replacements.Replacement {
	return rules.PackageImportTargetRule{}.Collect(rules.PackageImportTargetInput{
		File:    "package.json",
		Content: config.content,
		Targets: config.importTargets,
		Moves:   moves,
	})
}

func packageImportWarnings(config packageSpecifierConfig, moves []planning.FileMove) []adapterproto.Warning {
	return rules.PackageImportWarningRule{}.Collect(rules.PackageImportWarningInput{
		File:               "package.json",
		ConditionalImports: config.conditionalImports,
		Moves:              moves,
	})
}
