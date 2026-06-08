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
	TypeScriptPathAliasReason    = "javascript-typescript-path-alias"
	TypeScriptPathAliasRuleName  = "javascript.TypeScriptPathAliasRule"
	TypeScriptPathTargetReason   = "javascript-typescript-path-target"
	TypeScriptPathTargetRuleName = "javascript.TypeScriptPathTargetRule"
)

type TypeScriptPathAliasRule struct{}

type TypeScriptPathTarget struct {
	Target string
}

type TypeScriptPathAmbiguity struct {
	Alias   string
	Targets []string
}

type TypeScriptPathTargetInput struct {
	ProjectRoot string
	File        string
	Content     []byte
	PathBase    string
	Targets     []TypeScriptPathTarget
	Moves       []planning.FileMove
}

type TypeScriptPathWarningInput struct {
	ProjectRoot string
	File        string
	PathBase    string
	Ambiguities []TypeScriptPathAmbiguity
	Moves       []planning.FileMove
}

type TypeScriptPathTargetRule struct{}
type TypeScriptPathWarningRule struct{}

func (r TypeScriptPathAliasRule) Collect(mappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return PathAliasSpecifierRule{
		Reason: TypeScriptPathAliasReason,
		Rule:   TypeScriptPathAliasRuleName,
	}.Collect(mappings, moves)
}

func (r TypeScriptPathTargetRule) Collect(input TypeScriptPathTargetInput) []replacements.Replacement {
	targetRewrites := typeScriptPathTargetRewrites(input.ProjectRoot, input.PathBase, input.Targets, input.Moves)
	if len(targetRewrites) == 0 {
		return nil
	}

	compilerOptionsRange, ok := jsonconfig.ObjectPropertyRange(input.Content, "compilerOptions")
	if !ok {
		return nil
	}
	pathsRange, ok := jsonconfig.ObjectPropertyRangeIn(input.Content, compilerOptionsRange, "paths")
	if !ok {
		return nil
	}
	return jsonconfig.SingleStringArrayValueReplacements(input.File, input.Content, pathsRange, targetRewrites, TypeScriptPathTargetReason, TypeScriptPathTargetRuleName)
}

func typeScriptPathTargetRewrites(projectRoot string, pathBase string, targets []TypeScriptPathTarget, moves []planning.FileMove) map[string]string {
	rewrites := map[string]string{}
	for _, target := range targets {
		for _, move := range moves {
			oldReference, ok := TypeScriptTargetReference(projectRoot, pathBase, move.OldPath, target.Target)
			if !ok || oldReference != target.Target {
				continue
			}
			newReference, ok := TypeScriptTargetReference(projectRoot, pathBase, move.NewPath, target.Target)
			if !ok || oldReference == newReference {
				continue
			}
			rewrites[oldReference] = newReference
		}
	}
	return rewrites
}

func TypeScriptTargetReference(projectRoot string, pathBase string, targetPath string, existingStyle string) (string, bool) {
	if !IsJavaScriptModuleExtension(filepath.Ext(targetPath)) {
		return "", false
	}

	absoluteTarget := filepath.Join(projectRoot, filepath.FromSlash(targetPath))
	relative, err := filepath.Rel(pathBase, absoluteTarget)
	if err != nil {
		return "", false
	}
	relative = filepath.ToSlash(relative)
	if relative == "." || filepath.IsAbs(relative) || StartsWithParentTraversal(relative) {
		return "", false
	}

	relative = strings.TrimPrefix(relative, "./")
	if strings.HasPrefix(existingStyle, "./") {
		return "./" + relative, true
	}
	return relative, true
}

func (r TypeScriptPathWarningRule) Collect(input TypeScriptPathWarningInput) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, ambiguity := range input.Ambiguities {
		if !typeScriptAmbiguityReferencesMove(input.ProjectRoot, input.PathBase, ambiguity, input.Moves) {
			continue
		}
		warnings = append(warnings, adapterproto.Warning{
			File:    input.File,
			Message: "TypeScript path alias " + strconv.Quote(ambiguity.Alias) + " has multiple targets; skipped conservatively.",
		})
	}
	return warnings
}

func typeScriptAmbiguityReferencesMove(projectRoot string, pathBase string, ambiguity TypeScriptPathAmbiguity, moves []planning.FileMove) bool {
	for _, target := range ambiguity.Targets {
		if typeScriptTargetPatternReferencesMove(projectRoot, pathBase, target, moves) {
			return true
		}
	}
	return false
}

func typeScriptTargetPatternReferencesMove(projectRoot string, pathBase string, target string, moves []planning.FileMove) bool {
	if targetPrefix, ok := WildcardPrefix(target); ok {
		resolvedPrefix, ok, err := ResolveAliasTargetPrefix(projectRoot, pathBase, targetPrefix)
		if err != nil || !ok {
			return false
		}
		for _, move := range moves {
			if strings.HasPrefix(move.OldPath, resolvedPrefix) {
				return true
			}
		}
		return false
	}

	for _, move := range moves {
		oldReference, ok := TypeScriptTargetReference(projectRoot, pathBase, move.OldPath, target)
		if ok && oldReference == target {
			return true
		}
	}
	return false
}
