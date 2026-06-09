package rules

import (
	"strconv"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

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
