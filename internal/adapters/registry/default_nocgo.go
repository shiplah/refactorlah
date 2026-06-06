//go:build !cgo

package registry

import (
	"refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/golang"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func defaultAnalyzers() []Analyzer {
	return []Analyzer{
		goAnalyzer{analyzer: golang.NewAnalyzer()},
	}
}

type goAnalyzer struct {
	analyzer *golang.Analyzer
}

func (a goAnalyzer) Name() string {
	return "go"
}

func (a goAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (contract.AggregatedResponse, bool, error) {
	_ = scanConfig
	return a.analyzer.Analyze(projectRoot, plan)
}
