package rules

import (
	"path/filepath"
	"strings"

	"github.com/shiplah/refactorlah/internal/adapters/javascript/jsonconfig"
	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/replacements"
)

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
