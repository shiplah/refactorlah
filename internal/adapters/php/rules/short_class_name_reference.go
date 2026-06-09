//go:build cgo

package rules

import (
	"strings"

	"github.com/shiplah/refactorlah/internal/adapters/php/names"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/replacements"
)

const ShortClassNameReferenceRuleName = "php.ShortClassNameReferenceRule"

type ShortClassNameReferenceInput struct {
	File      string
	Source    []byte
	OldSymbol string
	NewSymbol string
}

type ShortClassNameReferenceRule struct{}

func (r ShortClassNameReferenceRule) Collect(document *treesitter.Document, input ShortClassNameReferenceInput) []replacements.Replacement {
	oldShort := names.Short(input.OldSymbol)
	newShort := names.Short(input.NewSymbol)
	if oldShort == "" || oldShort == newShort || !hasPlainUseImport(document, input.OldSymbol, oldShort) && !hasNamespaceLocalReference(document, input.OldSymbol, input.NewSymbol, oldShort) {
		return nil
	}

	skippedRanges := document.NodesByKind("namespace_use_declaration", "namespace_definition")
	classConstantNodes := document.NodesByKind("class_constant_access_expression")
	var result []replacements.Replacement
	for _, node := range document.NodesByKind("name", "qualified_name") {
		if node.Text != oldShort || treesitter.NodeInsideAnyRange(node, skippedRanges) {
			continue
		}
		if nodeInsideFullyQualifiedClassConstant(node, classConstantNodes) {
			continue
		}
		if !isSafeShortClassReference(input.Source, node.StartByte, node.EndByte) {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte,
			End:         node.EndByte,
			Replacement: newShort,
			Reason:      "php-short-class-name-reference",
			Rule:        ShortClassNameReferenceRuleName,
			Adapter:     "php",
		})
	}

	return result
}

func nodeInsideFullyQualifiedClassConstant(node treesitter.Node, classConstantNodes []treesitter.Node) bool {
	for _, classConstantNode := range classConstantNodes {
		if node.StartByte < classConstantNode.StartByte || node.EndByte > classConstantNode.EndByte {
			continue
		}

		nameStart, nameEnd, _, ok := classNameBeforeClassConstant(classConstantNode.Text)
		return ok && strings.Contains(classConstantNode.Text[nameStart:nameEnd], "\\")
	}

	return false
}

func hasNamespaceLocalReference(document *treesitter.Document, oldSymbol string, newSymbol string, oldShort string) bool {
	if declaredNamespace(document) != names.Namespace(oldSymbol) {
		return false
	}

	importedSymbol := existingNormalImports(document)[oldShort]
	if importedSymbol == "" {
		return true
	}

	oldSymbol = strings.TrimPrefix(oldSymbol, "\\")
	newSymbol = strings.TrimPrefix(newSymbol, "\\")
	return importedSymbol == oldSymbol || importedSymbol == newSymbol
}

func hasPlainUseImport(document *treesitter.Document, oldSymbol string, oldShort string) bool {
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		if !strings.Contains(node.Text, oldSymbol) {
			continue
		}
		if strings.Contains(strings.ToLower(node.Text), " as ") {
			continue
		}
		if !strings.Contains(node.Text, oldShort) {
			continue
		}
		return true
	}
	return false
}

func isSafeShortClassReference(source []byte, start int, end int) bool {
	if start > 0 && source[start-1] == '$' {
		return false
	}

	next := nextNonSpace(source, end)
	if next >= 0 && source[next] == ':' && next+1 < len(source) && source[next+1] == ':' {
		return true
	}
	if next >= 0 && source[next] == '$' {
		return true
	}

	previous := previousNonSpace(source, start-1)
	if previous >= 0 {
		switch source[previous] {
		case ':', '?', '|', '&', '(', ',':
			return true
		}
	}

	keyword := previousWord(source, start)
	return keyword == "new" ||
		keyword == "instanceof" ||
		keyword == "extends" ||
		keyword == "implements" ||
		keyword == "use" ||
		keyword == "catch"
}

func nextNonSpace(source []byte, index int) int {
	for index < len(source) {
		if !names.IsWhitespace(source[index]) {
			return index
		}
		index++
	}
	return -1
}

func previousNonSpace(source []byte, index int) int {
	for index >= 0 {
		if !names.IsWhitespace(source[index]) {
			return index
		}
		index--
	}
	return -1
}

func previousWord(source []byte, start int) string {
	index := previousNonSpace(source, start-1)
	if index < 0 || !names.IsIdentifierByte(source[index]) {
		return ""
	}

	end := index + 1
	for index >= 0 && names.IsIdentifierByte(source[index]) {
		index--
	}
	return string(source[index+1 : end])
}
