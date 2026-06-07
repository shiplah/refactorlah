//go:build cgo

package php

import (
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/config"
	"refactorlah/internal/planning"
)

func analyzePHP(t *testing.T, root string, plan planning.MovePlan) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	return analyzePHPWithConfig(t, root, plan, config.Config{})
}

func analyzePHPWithConfig(t *testing.T, root string, plan planning.MovePlan, scanConfig config.Config) (adapterproto.AggregatedResponse, bool, error) {
	t.Helper()
	return NewAnalyzer().Analyze(root, plan, scanConfig, scan.NewIndex(root, scanConfig))
}
