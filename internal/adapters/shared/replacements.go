package shared

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/planning"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

func MovePaths(plan planning.MovePlan) []string {
	paths := make([]string, 0, len(plan.Moves)*2)
	for _, move := range plan.Moves {
		paths = append(paths, move.OldPath, move.NewPath)
	}
	return paths
}

func ToAdapterReplacements(input []replacements.Replacement) []adapterproto.Replacement {
	output := make([]adapterproto.Replacement, 0, len(input))
	for _, replacement := range input {
		output = append(output, adapterproto.Replacement{
			File:        replacement.File,
			Start:       replacement.Start,
			End:         replacement.End,
			Replacement: replacement.Replacement,
			Reason:      replacement.Reason,
			Rule:        replacement.Rule,
			Adapter:     replacement.Adapter,
		})
	}
	return output
}
