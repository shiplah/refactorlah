//go:build cgo

package php

import (
	"os"
	"path/filepath"

	"refactorlah/internal/languages/php/syntax"
)

func (s *SymbolScanner) primarySymbolKind(projectRoot string, relativePath string, expectedShortName string) (string, bool) {
	source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", false
	}

	document, err := Parse(source)
	if err != nil {
		return "", false
	}
	defer document.Close()

	candidates := document.NodesByKind("class_declaration", "interface_declaration", "trait_declaration", "enum_declaration")
	var matchingKind string
	for _, candidate := range candidates {
		name := syntax.DeclarationName(candidate.Text)
		if name == expectedShortName {
			return phpSymbolKind(candidate.Kind), true
		}
		if matchingKind == "" {
			matchingKind = phpSymbolKind(candidate.Kind)
		}
	}

	if len(candidates) == 1 {
		return matchingKind, true
	}

	return "", false
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
