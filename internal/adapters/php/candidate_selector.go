//go:build cgo

package php

import (
	"sort"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/php/names"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
)

type CandidateFileSelector struct{}

func (s CandidateFileSelector) Query(mappings []adapterproto.SymbolMapping) scan.CandidateQuery {
	if len(mappings) == 0 {
		return scan.CandidateQuery{}
	}

	includePaths := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		includePaths = append(includePaths, mapping.OldPath)
	}

	return scan.CandidateQuery{
		Extensions:   []string{".php"},
		Needles:      candidateNeedles(mappings),
		IncludePaths: includePaths,
	}
}

func candidateNeedles(mappings []adapterproto.SymbolMapping) []string {
	index := map[string]bool{}
	for _, mapping := range mappings {
		index[mapping.OldSymbol] = true
		index[mapping.OldNamespace] = true
		index[names.Short(mapping.OldSymbol)] = true
		for needle := range literalHints(mapping) {
			index[needle] = true
		}
	}

	needles := make([]string, 0, len(index))
	for needle := range index {
		if needle != "" {
			needles = append(needles, needle)
		}
	}
	sort.Strings(needles)
	return needles
}
