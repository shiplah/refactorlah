//go:build !cgo

package javascript

import (
	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

func (a *Analyzer) collectBundlerAliasReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	return nil, nil, nil
}
