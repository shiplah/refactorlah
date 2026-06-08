//go:build cgo

package registry

import (
	"github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/golang"
	"github.com/NickSdot/refactorlah/internal/adapters/php"
	"github.com/NickSdot/refactorlah/internal/adapters/python"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/config"
	"github.com/NickSdot/refactorlah/internal/planning"
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

func (a goAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
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
