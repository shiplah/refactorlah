package rules

import (
	"path/filepath"
	"strconv"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/javascript/jsonconfig"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

const (
	PackageImportsReason         = "javascript-package-imports"
	PackageImportsRuleName       = "javascript.PackageImportsRule"
	PackageImportTargetReason    = "javascript-package-import-target"
	PackageImportTargetRuleName  = "javascript.PackageImportTargetRule"
	PackageSelfReferenceReason   = "javascript-package-self-reference"
	PackageSelfReferenceRuleName = "javascript.PackageSelfReferenceRule"
)

type PackageImportTarget struct {
	Target string
}

type PackageConditionalImport struct {
	Key     string
	Targets []string
}

type PackageImportAliasRule struct{}
type PackageImportTargetRule struct{}
type PackageImportWarningRule struct{}

type PackageImportTargetInput struct {
	File    string
	Content []byte
	Targets []PackageImportTarget
	Moves   []planning.FileMove
}

type PackageImportWarningInput struct {
	File               string
	ConditionalImports []PackageConditionalImport
	Moves              []planning.FileMove
}

func PackageSelfReferenceMappings(packageName string) []PathAliasMapping {
	aliasPrefix, ok := packageSelfReferenceAliasPrefix(packageName)
	if !ok {
		return nil
	}
	return []PathAliasMapping{{
		AliasPrefix: aliasPrefix,
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

func (r PackageImportAliasRule) Collect(importMappings []PathAliasMapping, selfReferenceMappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := PathAliasSpecifierRule{
		Reason: PackageImportsReason,
		Rule:   PackageImportsRuleName,
	}.Collect(importMappings, moves)
	rewrites = append(rewrites, PathAliasSpecifierRule{
		Reason: PackageSelfReferenceReason,
		Rule:   PackageSelfReferenceRuleName,
	}.Collect(selfReferenceMappings, moves)...)
	return rewrites
}

func (r PackageImportTargetRule) Collect(input PackageImportTargetInput) []replacements.Replacement {
	targetRewrites := packageImportTargetRewrites(input.Targets, input.Moves)
	if len(targetRewrites) == 0 {
		return nil
	}

	importsRange, ok := jsonconfig.ObjectPropertyRange(input.Content, "imports")
	if !ok {
		return nil
	}
	return jsonconfig.StringValueReplacements(input.File, input.Content, importsRange, targetRewrites, PackageImportTargetReason, PackageImportTargetRuleName)
}

func packageImportTargetRewrites(targets []PackageImportTarget, moves []planning.FileMove) map[string]string {
	rewrites := map[string]string{}
	for _, target := range targets {
		for _, move := range moves {
			if !IsJavaScriptModuleExtension(filepath.Ext(move.OldPath)) || !IsJavaScriptModuleExtension(filepath.Ext(move.NewPath)) {
				continue
			}

			oldTarget := "./" + filepath.ToSlash(move.OldPath)
			newTarget := "./" + filepath.ToSlash(move.NewPath)
			if target.Target != oldTarget || oldTarget == newTarget {
				continue
			}
			rewrites[oldTarget] = newTarget
		}
	}
	return rewrites
}

func (r PackageImportWarningRule) Collect(input PackageImportWarningInput) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, conditionalImport := range input.ConditionalImports {
		if !packageConditionalImportReferencesMove(conditionalImport, input.Moves) {
			continue
		}
		warnings = append(warnings, adapterproto.Warning{
			File:    input.File,
			Message: "Package imports entry " + strconv.Quote(conditionalImport.Key) + " uses conditional targets; skipped conservatively.",
		})
	}
	return warnings
}

func packageConditionalImportReferencesMove(conditionalImport PackageConditionalImport, moves []planning.FileMove) bool {
	for _, target := range conditionalImport.Targets {
		if packageTargetPatternReferencesMove(target, moves) {
			return true
		}
	}
	return false
}

func packageTargetPatternReferencesMove(target string, moves []planning.FileMove) bool {
	if targetPrefix, ok := WildcardPrefix(target); ok {
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
