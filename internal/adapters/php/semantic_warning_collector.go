//go:build cgo

package php

import (
	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/scan"
)

type SemanticWarningCollector struct {
	semanticHintScanner SemanticHintScanner
}

func NewSemanticWarningCollector() SemanticWarningCollector {
	return SemanticWarningCollector{
		semanticHintScanner: SemanticHintScanner{},
	}
}

func (c SemanticWarningCollector) Collect(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping, scanIndex *scan.Index) ([]adapterproto.Warning, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	phpFiles, err := scanIndex.Files(composerRoot, ".php")
	if err != nil {
		return nil, err
	}
	textFiles, err := scanIndex.Files(composerRoot, ".yaml", ".yml", ".xml", ".neon")
	if err != nil {
		return nil, err
	}

	return c.semanticHintScanner.Scan(projectRoot, phpFiles, textFiles, mappings)
}
