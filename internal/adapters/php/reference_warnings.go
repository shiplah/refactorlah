//go:build cgo

package php

import (
	"regexp"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/php/names"
	"github.com/NickSdot/refactorlah/internal/adapters/php/rules"
	"github.com/NickSdot/refactorlah/internal/parsing/treesitter"
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
	namespace := names.Namespace(symbol)
	short := names.Short(symbol)
	if namespace == "" || short == "" {
		return false
	}

	for _, prefix := range []string{namespace + "\\{", "\\" + namespace + "\\{"} {
		if strings.Contains(statement, prefix) && names.ContainsIdentifier(statement, short) {
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

func lineForByte(source []byte, offset int) int {
	if offset < 0 || offset > len(source) {
		return 0
	}
	return strings.Count(string(source[:offset]), "\n") + 1
}
