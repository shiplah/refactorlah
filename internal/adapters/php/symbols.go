//go:build cgo

package php

import (
	"path"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/planning"
)

type SymbolScanner struct {
	resolver Psr4NamespaceResolver
}

func NewSymbolScanner() *SymbolScanner {
	return &SymbolScanner{resolver: Psr4NamespaceResolver{}}
}

func (s *SymbolScanner) Scan(projectRoot string, psr4 Psr4Map, moves []planning.FileMove) ([]adapterproto.SymbolMapping, []adapterproto.Warning) {
	var mappings []adapterproto.SymbolMapping
	var warnings []adapterproto.Warning

	for _, move := range moves {
		if path.Ext(move.OldPath) != ".php" {
			continue
		}

		oldSymbol, oldOK := s.resolver.DeriveSymbol(psr4, move.OldPath)
		newSymbol, newOK := s.resolver.DeriveSymbol(psr4, move.NewPath)
		if !oldOK || !newOK {
			warnings = append(warnings, adapterproto.Warning{
				File:    move.OldPath,
				Message: "Moved PHP file is outside known PSR-4 roots; symbol mapping skipped.",
			})
			continue
		}

		topLevelMappings := s.topLevelConstantAndFunctionMappings(projectRoot, move, oldSymbol.Namespace, newSymbol.Namespace)
		symbolKind, ok, warningMessage := s.primarySymbolKind(projectRoot, move.OldPath, oldSymbol.ShortName)
		if !ok {
			if len(topLevelMappings) == 0 {
				warnings = append(warnings, adapterproto.Warning{
					File:    move.OldPath,
					Message: warningMessage,
				})
				continue
			}
		} else {
			mappings = append(mappings, adapterproto.SymbolMapping{
				Kind:         symbolKind,
				OldPath:      move.OldPath,
				NewPath:      move.NewPath,
				OldSymbol:    oldSymbol.Symbol,
				NewSymbol:    newSymbol.Symbol,
				OldNamespace: oldSymbol.Namespace,
				NewNamespace: newSymbol.Namespace,
				ShortName:    oldSymbol.ShortName,
			})
		}

		mappings = append(mappings, topLevelMappings...)
	}

	return mappings, warnings
}

func isPHPClassLikeKind(kind string) bool {
	switch kind {
	case "class", "interface", "trait", "enum":
		return true
	default:
		return false
	}
}
