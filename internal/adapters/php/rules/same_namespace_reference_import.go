//go:build cgo

package rules

import (
	"sort"
	"strings"

	"refactorlah/internal/adapters/php/names"
	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const SameNamespaceReferenceImportRuleName = "php.SameNamespaceReferenceImportRule"

type SameNamespaceReferenceImportInput struct {
	File     string
	Source   []byte
	Mappings []SymbolMappingReference
}

type SameNamespaceReferenceImportRule struct{}

func (r SameNamespaceReferenceImportRule) Collect(document *treesitter.Document, input SameNamespaceReferenceImportInput) []replacements.Replacement {
	namespace := declaredNamespace(document)
	if namespace == "" {
		return nil
	}

	declaredNames := declaredClassLikeNames(document)
	existingImports := existingNormalImports(document)
	skippedRanges := document.NodesByKind("namespace_use_declaration", "namespace_definition")
	plannedImports := map[string]string{}
	var result []replacements.Replacement

	for _, mapping := range input.Mappings {
		if names.Namespace(mapping.OldSymbol) != namespace || names.Namespace(mapping.NewSymbol) == namespace {
			continue
		}

		oldShort := names.Short(mapping.OldSymbol)
		if oldShort == "" || declaredNames[oldShort] {
			continue
		}

		foundReference := false
		for _, node := range document.NodesByKind("name", "qualified_name") {
			if node.Text != oldShort || treesitter.NodeInsideAnyRange(node, skippedRanges) {
				continue
			}
			if !isSafeShortClassReference(input.Source, node.StartByte, node.EndByte) {
				continue
			}
			if existingImports[oldShort] != "" {
				continue
			}

			foundReference = true
		}

		if !foundReference || existingImports[oldShort] != "" {
			continue
		}
		plannedImports[oldShort] = mapping.NewSymbol
	}

	if len(plannedImports) == 0 {
		return result
	}
	if insertion, ok := sameNamespaceReferenceImportInsertion(document, plannedImports, input.File); ok {
		result = append(result, insertion)
	}

	return result
}

func sameNamespaceReferenceImportInsertion(document *treesitter.Document, imports map[string]string, file string) (replacements.Replacement, bool) {
	symbols := make([]string, 0, len(imports))
	for _, symbol := range imports {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	rendered := renderUseStatements(symbols)
	useStatements := document.NodesByKind("namespace_use_declaration")
	if len(useStatements) > 0 {
		lastUse := useStatements[len(useStatements)-1]
		return replacements.Replacement{
			File:        file,
			Start:       lastUse.EndByte,
			End:         lastUse.EndByte,
			Replacement: "\n" + rendered,
			Reason:      "php-same-namespace-reference-import",
			Rule:        SameNamespaceReferenceImportRuleName,
			Adapter:     "php",
		}, true
	}

	for _, namespace := range document.NodesByKind("namespace_definition") {
		semicolon := strings.Index(namespace.Text, ";")
		if semicolon < 0 {
			return replacements.Replacement{}, false
		}

		offset := namespace.StartByte + semicolon + 1
		return replacements.Replacement{
			File:        file,
			Start:       offset,
			End:         offset,
			Replacement: "\n\n" + rendered,
			Reason:      "php-same-namespace-reference-import",
			Rule:        SameNamespaceReferenceImportRuleName,
			Adapter:     "php",
		}, true
	}

	return replacements.Replacement{}, false
}
