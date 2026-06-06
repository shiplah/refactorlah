//go:build cgo

package rules

import (
	"strings"

	"refactorlah/internal/adapters/php/names"
	"refactorlah/internal/parsing/treesitter"
)

type SymbolReferenceInput struct {
	File         string
	Source       []byte
	OldSymbol    string
	NewSymbol    string
	OldNamespace string
	NewNamespace string
}

type SymbolMappingReference struct {
	OldSymbol string
	NewSymbol string
}

func declaredNamespace(document *treesitter.Document) string {
	for _, node := range document.NodesByKind("namespace_definition") {
		text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(node.Text), "namespace"))
		if end := strings.IndexAny(text, ";{"); end >= 0 {
			text = text[:end]
		}
		return strings.TrimSpace(text)
	}

	return ""
}

func existingNormalImports(document *treesitter.Document) map[string]string {
	imports := map[string]string{}
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		if strings.Contains(strings.ToLower(node.Text), " as ") {
			continue
		}

		importedSymbol := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(node.Text), ";"), "use"))
		if importedSymbol == "" || strings.Contains(importedSymbol, ",") || strings.Contains(importedSymbol, "{") {
			continue
		}

		imports[names.Short(importedSymbol)] = importedSymbol
	}

	return imports
}

func importedShortReplacement(document *treesitter.Document, oldSymbol string, newSymbol string, reference string) (string, bool) {
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		if !strings.Contains(node.Text, oldSymbol) && !strings.Contains(node.Text, newSymbol) {
			continue
		}
		if strings.Contains(strings.ToLower(node.Text), " as ") {
			continue
		}
		if reference != names.Short(oldSymbol) {
			continue
		}

		return names.Short(newSymbol), true
	}

	return "", false
}
