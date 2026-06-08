//go:build !cgo

package javascript

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/planning"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

func (a *Analyzer) collectBundlerAliasReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	return nil, nil, nil
}
