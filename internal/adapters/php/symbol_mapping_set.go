//go:build cgo

package php

import (
	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/php/rules"
)

type SymbolMappingSet struct {
	mappings   []adapterproto.SymbolMapping
	byOldPath  map[string]adapterproto.SymbolMapping
	references []rules.SymbolMappingReference
}

func NewSymbolMappingSet(mappings []adapterproto.SymbolMapping) SymbolMappingSet {
	set := SymbolMappingSet{
		mappings:   append([]adapterproto.SymbolMapping(nil), mappings...),
		byOldPath:  map[string]adapterproto.SymbolMapping{},
		references: make([]rules.SymbolMappingReference, 0, len(mappings)),
	}

	for _, mapping := range mappings {
		set.byOldPath[mapping.OldPath] = mapping
		set.references = append(set.references, rules.SymbolMappingReference{
			OldSymbol: mapping.OldSymbol,
			NewSymbol: mapping.NewSymbol,
		})
	}

	return set
}

func (s SymbolMappingSet) Len() int {
	return len(s.mappings)
}

func (s SymbolMappingSet) All() []adapterproto.SymbolMapping {
	return append([]adapterproto.SymbolMapping(nil), s.mappings...)
}

func (s SymbolMappingSet) MovedFile(path string) (adapterproto.SymbolMapping, bool) {
	mapping, ok := s.byOldPath[path]
	return mapping, ok
}

func (s SymbolMappingSet) References() []rules.SymbolMappingReference {
	return append([]rules.SymbolMappingReference(nil), s.references...)
}
