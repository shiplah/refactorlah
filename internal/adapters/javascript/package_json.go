package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

const (
	packageImportsReason       = "javascript-package-imports"
	packageImportsRule         = "javascript.PackageImportsRule"
	packageImportTargetReason  = "javascript-package-import-target"
	packageImportTargetRule    = "javascript.PackageImportTargetRule"
	packageSelfReferenceReason = "javascript-package-self-reference"
	packageSelfReferenceRule   = "javascript.PackageSelfReferenceRule"
)

type packageSpecifierConfig struct {
	content               []byte
	importTargets         []packageImportTarget
	conditionalImports    []packageConditionalImport
	importMappings        []pathAliasMapping
	selfReferenceMappings []pathAliasMapping
}

type packageImportTarget struct {
	target string
}

type packageConditionalImport struct {
	key     string
	targets []string
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

func buildPackageImportTargets(imports map[string]json.RawMessage) []packageImportTarget {
	if len(imports) == 0 {
		return nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var targets []packageImportTarget
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
		if !isJavaScriptModuleExtension(filepath.Ext(target)) {
			continue
		}

		targets = append(targets, packageImportTarget{target: target})
	}
	return targets
}

func buildPackageConditionalImports(imports map[string]json.RawMessage) []packageConditionalImport {
	if len(imports) == 0 {
		return nil
	}

	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)

	var conditionalImports []packageConditionalImport
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
		conditionalImports = append(conditionalImports, packageConditionalImport{
			key:     importKey,
			targets: targets,
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

func packageImportTargetReplacements(config packageSpecifierConfig, moves []planning.FileMove) []replacements.Replacement {
	targetRewrites := packageImportTargetRewrites(config.importTargets, moves)
	if len(targetRewrites) == 0 {
		return nil
	}

	importsRange, ok := jsonObjectPropertyRange(config.content, "imports")
	if !ok {
		return nil
	}
	return jsonObjectStringValueReplacements("package.json", config.content, importsRange, targetRewrites, packageImportTargetReason, packageImportTargetRule)
}

func packageImportTargetRewrites(targets []packageImportTarget, moves []planning.FileMove) map[string]string {
	rewrites := map[string]string{}
	for _, target := range targets {
		for _, move := range moves {
			if !isJavaScriptModuleExtension(filepath.Ext(move.OldPath)) || !isJavaScriptModuleExtension(filepath.Ext(move.NewPath)) {
				continue
			}

			oldTarget := "./" + filepath.ToSlash(move.OldPath)
			newTarget := "./" + filepath.ToSlash(move.NewPath)
			if target.target != oldTarget || oldTarget == newTarget {
				continue
			}
			rewrites[oldTarget] = newTarget
		}
	}
	return rewrites
}

func packageImportWarnings(config packageSpecifierConfig, moves []planning.FileMove) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, conditionalImport := range config.conditionalImports {
		if !packageConditionalImportReferencesMove(conditionalImport, moves) {
			continue
		}
		warnings = append(warnings, adapterproto.Warning{
			File:    "package.json",
			Message: "Package imports entry " + strconv.Quote(conditionalImport.key) + " uses conditional targets; skipped conservatively.",
		})
	}
	return warnings
}

func packageConditionalImportReferencesMove(conditionalImport packageConditionalImport, moves []planning.FileMove) bool {
	for _, target := range conditionalImport.targets {
		if packageTargetPatternReferencesMove(target, moves) {
			return true
		}
	}
	return false
}

func packageTargetPatternReferencesMove(target string, moves []planning.FileMove) bool {
	if targetPrefix, ok := wildcardPrefix(target); ok {
		resolvedPrefix := strings.TrimPrefix(targetPrefix, "./")
		for _, move := range moves {
			if strings.HasPrefix(move.OldPath, resolvedPrefix) {
				return true
			}
		}
		return false
	}

	for _, move := range moves {
		oldTarget := "./" + filepath.ToSlash(move.OldPath)
		if target == oldTarget {
			return true
		}
	}
	return false
}
