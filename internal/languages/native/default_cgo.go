//go:build cgo

package native

import (
	"refactorlah/internal/adapters"
	"refactorlah/internal/config"
	"refactorlah/internal/languages/golang"
	"refactorlah/internal/languages/php"
	"refactorlah/internal/languages/python"
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

func (a goAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, bool, error) {
	_ = scanConfig
	return a.analyzer.Analyze(projectRoot, plan)
}

type phpAnalyzer struct {
	analyzer *php.Analyzer
}

func (a phpAnalyzer) Name() string {
	return "php"
}

func (a phpAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, bool, error) {
	return a.analyzer.AnalyzeWithConfig(projectRoot, plan, scanConfig)
}

type pythonAnalyzer struct {
	analyzer *python.Analyzer
}

func (a pythonAnalyzer) Name() string {
	return "python"
}

func (a pythonAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config) (adapters.AggregatedResponse, bool, error) {
	return a.analyzer.AnalyzeWithConfig(projectRoot, plan, scanConfig)
}
