package registry

import (
	"github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

type Analyzer interface {
	Name() string
	Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error)
}

type Registry struct {
	analyzers []Analyzer
}

func NewRegistry() *Registry {
	return &Registry{analyzers: defaultAnalyzers()}
}

func EmptyRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (contract.AggregatedResponse, []string, error) {
	var merged contract.AggregatedResponse
	var names []string
	scanIndex := scan.NewIndex(projectRoot, scanConfig)

	for _, analyzer := range r.analyzers {
		response, relevant, err := analyzer.Analyze(projectRoot, plan, scanConfig, scanIndex)
		if err != nil {
			return contract.AggregatedResponse{}, names, err
		}
		if !relevant {
			continue
		}

		merged = merge(merged, response)
		names = append(names, analyzer.Name())
	}

	return merged, names, nil
}

func merge(left contract.AggregatedResponse, right contract.AggregatedResponse) contract.AggregatedResponse {
	left.SymbolMappings = append(left.SymbolMappings, right.SymbolMappings...)
	left.PathMappings = append(left.PathMappings, right.PathMappings...)
	left.Replacements = append(left.Replacements, right.Replacements...)
	left.Warnings = append(left.Warnings, right.Warnings...)
	left.Checks = append(left.Checks, right.Checks...)
	return left
}
