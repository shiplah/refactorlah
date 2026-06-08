package rules

import (
	"path/filepath"
	"strconv"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/planning"
)

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
