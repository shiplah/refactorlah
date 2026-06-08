package registry

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/config"
	"github.com/NickSdot/refactorlah/internal/planning"
)

func TestRegistrySharesOneScanIndexAcrossAnalyzers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	var firstIndex *scan.Index
	secondSawSameIndex := false

	registry := &Registry{analyzers: []Analyzer{
		capturingAnalyzer{
			name: "first",
			onAnalyze: func(scanIndex *scan.Index) {
				firstIndex = scanIndex
			},
		},
		capturingAnalyzer{
			name: "second",
			onAnalyze: func(scanIndex *scan.Index) {
				secondSawSameIndex = firstIndex != nil && firstIndex == scanIndex
			},
		},
	}}

	_, names, err := registry.Analyze(root, planning.MovePlan{}, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if !secondSawSameIndex {
		t.Fatal("expected analyzers to share one scan index")
	}
	if len(names) != 2 {
		t.Fatalf("expected both analyzers to be reported relevant, got %#v", names)
	}
}

type capturingAnalyzer struct {
	name      string
	onAnalyze func(*scan.Index)
}

func (a capturingAnalyzer) Name() string {
	return a.name
}

func (a capturingAnalyzer) Analyze(projectRoot string, plan planning.MovePlan, scanConfig config.Config, scanIndex *scan.Index) (contract.AggregatedResponse, bool, error) {
	a.onAnalyze(scanIndex)
	return contract.AggregatedResponse{}, true, nil
}
