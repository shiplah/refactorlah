//go:build cgo

package registry

import (
	"refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/golang"
	"refactorlah/internal/adapters/php"
	"refactorlah/internal/adapters/python"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func defaultAnalyzers() []Analyzer {
	return []Analyzer{
		goAnalyzer{analyzer: golang.NewAnalyzer()},
		phpAnalyzer{analyzer: php.NewAnalyzer()},
		pythonAnalyzer{analyzer: python.NewAnalyzer()},
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

type phpAnalyzer struct {
	analyzer *php.Analyzer
}

func (a phpAnalyzer) Name() string {
	return "php"
}

func (a phpAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.AnalyzeWithConfig(projectRoot, plan, scanConfig)
}

type pythonAnalyzer struct {
	analyzer *python.Analyzer
}

func (a pythonAnalyzer) Name() string {
	return "python"
}

func (a pythonAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.AnalyzeWithConfig(projectRoot, plan, scanConfig)
}
