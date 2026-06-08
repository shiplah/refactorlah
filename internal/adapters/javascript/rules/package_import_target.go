package rules

import (
	"path/filepath"

	"refactorlah/internal/adapters/javascript/jsonconfig"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

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
