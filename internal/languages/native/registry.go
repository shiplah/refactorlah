package native

import (
	"refactorlah/internal/adapters"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

type Analyzer interface {
	Name() string
	Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, bool, error)
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

func (r *Registry) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, []string, error) {
	var merged adapters.AggregatedResponse
	var names []string

	for _, analyzer := range r.analyzers {
		response, relevant, err := analyzer.Analyze(projectRoot, plan, scanConfig)
		if err != nil {
			return adapters.AggregatedResponse{}, names, err
		}
		if !relevant {
			continue
		}

		merged = merge(merged, response)
		names = append(names, analyzer.Name())
	}

	return merged, names, nil
}

func merge(left adapters.AggregatedResponse, right adapters.AggregatedResponse) adapters.AggregatedResponse {
	left.SymbolMappings = append(left.SymbolMappings, right.SymbolMappings...)
	left.PathMappings = append(left.PathMappings, right.PathMappings...)
	left.Replacements = append(left.Replacements, right.Replacements...)
	left.Warnings = append(left.Warnings, right.Warnings...)
	return left
}
