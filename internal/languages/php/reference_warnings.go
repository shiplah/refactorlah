//go:build cgo

package php

import (
	"regexp"
	"strings"

	adapterproto "refactorlah/internal/adapters"
	"refactorlah/internal/languages/php/rules"
	"refactorlah/internal/parsing/treesitter"
)

func collectReferenceWarnings(document *treesitter.Document, file string, source []byte, mappings []rules.SymbolMappingReference) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	warnings = append(warnings, collectGroupUseWarnings(document, file, source, mappings)...)
	warnings = append(warnings, collectStringLiteralSymbolWarnings(file, source, mappings)...)
	return warnings
}

func collectGroupUseWarnings(document *treesitter.Document, file string, source []byte, mappings []rules.SymbolMappingReference) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		if !strings.Contains(node.Text, "{") {
			continue
		}

		for _, mapping := range mappings {
			if !groupUseReferencesSymbol(node.Text, mapping.OldSymbol) {
				continue
			}

			warnings = append(warnings, adapterproto.Warning{
				File:    file,
				Line:    lineForByte(source, node.StartByte),
				Message: "Group use statement references a moved symbol; skipped conservatively.",
			})
			break
		}
	}
	return warnings
}

func groupUseReferencesSymbol(statement string, symbol string) bool {
	symbol = strings.TrimPrefix(symbol, "\\")
	namespace := namespaceOfSymbol(symbol)
	short := shortSymbolName(symbol)
	if namespace == "" || short == "" {
		return false
	}

	for _, prefix := range []string{namespace + "\\{", "\\" + namespace + "\\{"} {
		if strings.Contains(statement, prefix) && containsIdentifier(statement, short) {
			return true
		}
	}
	return false
}

func collectStringLiteralSymbolWarnings(file string, source []byte, mappings []rules.SymbolMappingReference) []adapterproto.Warning {
	content := string(source)
	var warnings []adapterproto.Warning
	for _, mapping := range mappings {
		oldSymbol := strings.TrimPrefix(mapping.OldSymbol, "\\")
		if oldSymbol == "" {
			continue
		}

		for _, offset := range stringLiteralSymbolOffsets(content, oldSymbol) {
			warnings = append(warnings, adapterproto.Warning{
				File:    file,
				Line:    lineForByte(source, offset),
				Message: "String literal references a moved PHP symbol; not changed.",
			})
		}
	}
	return warnings
}

func stringLiteralSymbolOffsets(content string, symbol string) []int {
	quotedSymbol := regexp.QuoteMeta(symbol)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`'[^'\r\n]*` + quotedSymbol + `[^'\r\n]*'`),
		regexp.MustCompile(`"[^"\r\n]*` + quotedSymbol + `[^"\r\n]*"`),
	}

	seen := map[int]bool{}
	var offsets []int
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringIndex(content, -1) {
			if match[0] < 0 || seen[match[0]] {
				continue
			}
			seen[match[0]] = true
			offsets = append(offsets, match[0])
		}
	}
	return offsets
}

func namespaceOfSymbol(symbol string) string {
	index := strings.LastIndex(symbol, "\\")
	if index < 0 {
		return ""
	}
	return symbol[:index]
}

func containsIdentifier(content string, identifier string) bool {
	offset := 0
	for {
		index := strings.Index(content[offset:], identifier)
		if index < 0 {
			return false
		}

		start := offset + index
		end := start + len(identifier)
		if isIdentifierBoundary(content, start-1) && isIdentifierBoundary(content, end) {
			return true
		}
		offset = end
	}
}

func isIdentifierBoundary(content string, index int) bool {
	if index < 0 || index >= len(content) {
		return true
	}
	return !isPHPIdentifierByteForWarning(content[index])
}

func isPHPIdentifierByteForWarning(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func lineForByte(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 0
	}
	return strings.Count(string(source[:offset]), "\n") + 1
}
