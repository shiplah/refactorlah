//go:build cgo

package rules

import (
	"strings"

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

		imports[phpShortName(importedSymbol)] = importedSymbol
	}

	return imports
}

func namespaceOf(symbol string) string {
	index := strings.LastIndex(symbol, "\\")
	if index < 0 {
		return ""
	}
	return symbol[:index]
}

func importedShortReplacement(document *treesitter.Document, oldSymbol string, newSymbol string, reference string) (string, bool) {
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		if !strings.Contains(node.Text, oldSymbol) && !strings.Contains(node.Text, newSymbol) {
			continue
		}
		if strings.Contains(strings.ToLower(node.Text), " as ") {
			continue
		}
		if reference != phpShortName(oldSymbol) {
			continue
		}

		return phpShortName(newSymbol), true
	}

	return "", false
}
