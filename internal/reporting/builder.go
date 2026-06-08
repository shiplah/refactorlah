package reporting

import (
	"sort"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/planning"
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) MoveReports(plan planning.MovePlan) []MoveReport {
	reports := make([]MoveReport, 0, len(plan.Moves))
	for _, move := range plan.Moves {
		reports = append(reports, MoveReport{
			OldPath: move.OldPath,
			NewPath: move.NewPath,
			Tracked: move.Tracked,
			Mover:   move.Mover,
		})
	}
	return reports
}

func (b *Builder) SymbolMappings(items []adapterproto.SymbolMapping) []SymbolMapping {
	result := make([]SymbolMapping, 0, len(items))
	for _, item := range items {
		result = append(result, SymbolMapping{
			Kind:      item.Kind,
			OldPath:   item.OldPath,
			NewPath:   item.NewPath,
			OldSymbol: item.OldSymbol,
			NewSymbol: item.NewSymbol,
		})
	}
	return result
}

func (b *Builder) PathMappings(items []adapterproto.PathMapping) []PathMapping {
	result := make([]PathMapping, 0, len(items))
	for _, item := range items {
		result = append(result, PathMapping{
			Kind:         item.Kind,
			OldPath:      item.OldPath,
			NewPath:      item.NewPath,
			OldReference: item.OldReference,
			NewReference: item.NewReference,
		})
	}
	return result
}

func (b *Builder) EditedFiles(replacements []adapterproto.Replacement) []EditedFile {
	counts := map[string]int{}
	for _, replacement := range replacements {
		counts[replacement.File]++
	}

	files := make([]EditedFile, 0, len(counts))
	for file, count := range counts {
		files = append(files, EditedFile{File: file, Replacements: count})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].File < files[j].File
	})
	return files
}

func (b *Builder) Replacements(items []adapterproto.Replacement) []ReplacementReport {
	result := make([]ReplacementReport, 0, len(items))
	for _, item := range items {
		result = append(result, ReplacementReport{
			File:        item.File,
			Start:       item.Start,
			End:         item.End,
			Reason:      item.Reason,
			Rule:        item.Rule,
			Adapter:     item.Adapter,
			Replacement: item.Replacement,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].File == result[j].File {
			return result[i].Start < result[j].Start
		}
		return result[i].File < result[j].File
	})
	return result
}

func (b *Builder) RuleResults(items []adapterproto.Replacement) []RuleResult {
	counts := map[string]int{}
	for _, item := range items {
		rule := item.Rule
		if rule == "" {
			rule = item.Reason
		}
		counts[rule]++
	}

	result := make([]RuleResult, 0, len(counts))
	for rule, count := range counts {
		result = append(result, RuleResult{
			Rule:         rule,
			Replacements: count,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Rule < result[j].Rule
	})
	return result
}
