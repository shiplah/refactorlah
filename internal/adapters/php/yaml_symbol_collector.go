//go:build cgo

package php

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/php/symfony/twig"
	"github.com/NickSdot/refactorlah/internal/adapters/scan"
	"github.com/NickSdot/refactorlah/internal/adapters/shared"
)

type YamlSymbolCollector struct {
	componentNamespaceScanner twig.ComponentNamespaceScanner
}

func NewYamlSymbolCollector() YamlSymbolCollector {
	return YamlSymbolCollector{
		componentNamespaceScanner: twig.ComponentNamespaceScanner{},
	}
}

func (c YamlSymbolCollector) Collect(projectRoot string, composerRoot string, mappings []adapterproto.SymbolMapping, scanIndex *scan.Index) ([]adapterproto.Replacement, error) {
	if len(mappings) == 0 {
		return nil, nil
	}

	yamlFiles, err := scanIndex.Files(composerRoot, ".yaml", ".yml")
	if err != nil {
		return nil, err
	}
	componentNamespaceReplacements, err := c.componentNamespaceScanner.Scan(projectRoot, yamlFiles, mappings)
	if err != nil {
		return nil, err
	}

	return shared.ToAdapterReplacements(componentNamespaceReplacements), nil
}
