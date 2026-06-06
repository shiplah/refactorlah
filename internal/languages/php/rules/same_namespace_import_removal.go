//go:build cgo

package rules

import (
	"strings"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const SameNamespaceImportRemovalRuleName = "php.SameNamespaceImportRemovalRule"

type SameNamespaceImportRemovalInput struct {
	File         string
	Source       []byte
	NewNamespace string
	Mappings     []SymbolMappingReference
}

type SameNamespaceImportRemovalRule struct{}

func (r SameNamespaceImportRemovalRule) Collect(document *treesitter.Document, input SameNamespaceImportRemovalInput) []replacements.Replacement {
	if input.NewNamespace == "" {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		importedSymbol, ok := plainImportedSymbol(node.Text)
		if !ok {
			continue
		}
		if !importBecomesSameNamespace(importedSymbol, input.NewNamespace, input.Mappings) {
			continue
		}

		start, end := wholeLineRange(input.Source, node.StartByte, node.EndByte)
		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       start,
			End:         end,
			Replacement: "",
			Reason:      "php-same-namespace-import-removal",
			Rule:        SameNamespaceImportRemovalRuleName,
			Adapter:     "php",
		})
	}

	return result
}

func plainImportedSymbol(useStatement string) (string, bool) {
	text := strings.TrimSpace(useStatement)
	if !strings.HasPrefix(text, "use ") || !strings.HasSuffix(text, ";") {
		return "", false
	}
	if strings.Contains(strings.ToLower(text), " as ") || strings.Contains(text, "{") || strings.Contains(text, ",") {
		return "", false
	}

	symbol := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "use "), ";"))
	return symbol, symbol != ""
}

func importBecomesSameNamespace(importedSymbol string, newNamespace string, mappings []SymbolMappingReference) bool {
	if namespaceOf(importedSymbol) == newNamespace {
		return true
	}

	for _, mapping := range mappings {
		if mapping.OldSymbol == importedSymbol || mapping.NewSymbol == importedSymbol {
			return namespaceOf(mapping.NewSymbol) == newNamespace
		}
	}

	return false
}

func wholeLineRange(source []byte, start int, end int) (int, int) {
	for start > 0 && source[start-1] != '\n' && source[start-1] != '\r' {
		start--
	}
	for end < len(source) && (source[end] == '\n' || source[end] == '\r') {
		end++
	}

	return start, end
}
