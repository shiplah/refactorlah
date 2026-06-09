package golang

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

func analyzeGo(t *testing.T, root string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	scanConfig := config.Config{}
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}

func analyzeGoWithConfig(t *testing.T, root string, plan planning.MovePlan, scanConfig config.Config) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}
