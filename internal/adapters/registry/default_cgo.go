//go:build cgo

package registry

import (
	"github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/golang"
	"github.com/shiplah/refactorlah/internal/adapters/javascript"
	"github.com/shiplah/refactorlah/internal/adapters/php"
	"github.com/shiplah/refactorlah/internal/adapters/python"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

func defaultAnalyzers() []Analyzer {
	return []Analyzer{
		goAnalyzer{analyzer: golang.NewAnalyzer()},
		javascriptAnalyzer{analyzer: javascript.NewAnalyzer()},
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

func (a goAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.Analyze(projectRoot, plan, scanConfig, scanIndex)
}

type javascriptAnalyzer struct {
	analyzer *javascript.Analyzer
}

func (a javascriptAnalyzer) Name() string {
	return "javascript"
}

func (a javascriptAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.Analyze(projectRoot, plan, scanConfig, scanIndex)
}

type phpAnalyzer struct {
	analyzer *php.Analyzer
}

func (a phpAnalyzer) Name() string {
	return "php"
}

func (a phpAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.Analyze(projectRoot, plan, scanConfig, scanIndex)
}

type pythonAnalyzer struct {
	analyzer *python.Analyzer
}

func (a pythonAnalyzer) Name() string {
	return "python"
}

func (a pythonAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
	return a.analyzer.Analyze(projectRoot, plan, scanConfig, scanIndex)
}
