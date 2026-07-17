//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"strings"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/php/names"
	"github.com/shiplah/refactorlah/internal/adapters/php/syntax"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/planning"
)

func (s *SymbolScanner) primarySymbolKind(projectRoot string, relativePath string, expectedShortName string) (string, bool, string) {
	source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", false, "PHP file could not be read; symbol mapping skipped."
	}

	document, err := Parse(source)
	if err != nil {
		return "", false, "Moved PHP file top-level symbol could not be mapped deterministically; symbol mapping skipped."
	}
	defer document.Close()

	candidates := topLevelSymbolCandidates(document.NodesByKind("class_declaration", "interface_declaration", "trait_declaration", "enum_declaration"))
	for _, candidate := range candidates {
		name := syntax.DeclarationName(candidate.Text)
		if name == expectedShortName {
			return phpSymbolKind(candidate.Kind), true, ""
		}
	}

	if len(candidates) == 1 {
		return "", false, "Top-level symbol does not match deterministic PSR-4 filename; symbol mapping skipped."
	}

	return "", false, "Top-level symbol could not be matched deterministically; symbol mapping skipped."
}

func topLevelSymbolCandidates(candidates []treesitter.Node) []treesitter.Node {
	var topLevel []treesitter.Node
	for _, candidate := range candidates {
		if isTopLevelSymbolCandidate(candidate) {
			topLevel = append(topLevel, candidate)
		}
	}

	return topLevel
}

func isTopLevelSymbolCandidate(candidate treesitter.Node) bool {
	switch candidate.ParentKind() {
	case "program", "namespace_definition":
		return true
	default:
		return false
	}
}

func phpSymbolKind(nodeKind string) string {
	switch nodeKind {
	case "interface_declaration":
		return "interface"
	case "trait_declaration":
		return "trait"
	case "enum_declaration":
		return "enum"
	default:
		return "class"
	}
}

func (s *SymbolScanner) topLevelConstantAndFunctionMappings(projectRoot string, move planning.FileMove, oldNamespace string, newNamespace string) []adapterproto.SymbolMapping {
	if oldNamespace == "" || newNamespace == "" {
		return nil
	}

	source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(move.OldPath)))
	if err != nil {
		return nil
	}

	document, err := Parse(source)
	if err != nil {
		return nil
	}
	defer document.Close()

	var mappings []adapterproto.SymbolMapping
	for _, node := range document.NodesByKind("const_element") {
		if !isTopLevelConstantElement(node) {
			continue
		}
		name := constantElementName(node.Text)
		if name == "" {
			continue
		}
		mappings = append(mappings, topLevelSymbolMapping("constant", move, oldNamespace, newNamespace, name))
	}

	for _, node := range document.NodesByKind("function_definition") {
		if !isTopLevelSymbolCandidate(node) {
			continue
		}
		name := functionDefinitionName(node.Text)
		if name == "" {
			continue
		}
		mappings = append(mappings, topLevelSymbolMapping("function", move, oldNamespace, newNamespace, name))
	}

	return mappings
}

func topLevelSymbolMapping(kind string, move planning.FileMove, oldNamespace string, newNamespace string, shortName string) adapterproto.SymbolMapping {
	return adapterproto.SymbolMapping{
		Kind:         kind,
		OldPath:      move.OldPath,
		NewPath:      move.NewPath,
		OldSymbol:    oldNamespace + "\\" + shortName,
		NewSymbol:    newNamespace + "\\" + shortName,
		OldNamespace: oldNamespace,
		NewNamespace: newNamespace,
		ShortName:    shortName,
	}
}

func isTopLevelConstantElement(candidate treesitter.Node) bool {
	if candidate.ParentKind() != "const_declaration" {
		return false
	}

	hasTopLevelParent := false
	for _, kind := range candidate.AncestorKinds {
		switch kind {
		case "class_declaration", "interface_declaration", "trait_declaration", "enum_declaration":
			return false
		case "program", "namespace_definition":
			hasTopLevelParent = true
		}
	}

	return hasTopLevelParent
}

func constantElementName(text string) string {
	text = strings.TrimSpace(text)
	end := 0
	for end < len(text) && names.IsIdentifierByte(text[end]) {
		end++
	}
	return text[:end]
}

func functionDefinitionName(text string) string {
	index := 0
	for {
		found := indexOfSymbolWord(text[index:], "function")
		if found < 0 {
			return ""
		}

		start := index + found + len("function")
		for start < len(text) && names.IsWhitespace(text[start]) {
			start++
		}
		if start < len(text) && text[start] == '&' {
			start++
			for start < len(text) && names.IsWhitespace(text[start]) {
				start++
			}
		}

		end := start
		for end < len(text) && names.IsIdentifierByte(text[end]) {
			end++
		}
		if end > start {
			return text[start:end]
		}

		index = start
	}
}

func indexOfSymbolWord(text string, word string) int {
	for index := 0; index+len(word) <= len(text); index++ {
		if text[index:index+len(word)] != word {
			continue
		}

		before := index - 1
		after := index + len(word)
		if (before < 0 || !names.IsIdentifierByte(text[before])) && (after >= len(text) || !names.IsIdentifierByte(text[after])) {
			return index
		}
	}

	return -1
}
