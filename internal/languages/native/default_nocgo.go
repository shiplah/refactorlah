//go:build !cgo

package native

import (
	"refactorlah/internal/adapters"
	"refactorlah/internal/config"
	"refactorlah/internal/languages/golang"
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

func (a goAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, bool, error) {
	_ = scanConfig
	return a.analyzer.Analyze(projectRoot, plan)
}
