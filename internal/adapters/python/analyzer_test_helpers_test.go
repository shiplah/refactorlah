//go:build cgo

package python

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
	"github.com/shiplah/refactorlah/internal/config"
	"github.com/shiplah/refactorlah/internal/planning"
)

func analyzePython(t *testing.T, root string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	return analyzePythonWithConfig(t, root, plan, config.Config{})
}

func analyzePythonWithConfig(t *testing.T, root string, plan planning.MovePlan, scanConfig config.Config) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}
