//go:build cgo

package php

import (
	"os"
	"path/filepath"

	"github.com/shiplah/refactorlah/internal/adapters/php/syntax"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
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
