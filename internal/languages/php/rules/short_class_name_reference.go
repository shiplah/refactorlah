//go:build cgo

package rules

import (
	"strings"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
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
	oldShort := phpShortName(input.OldSymbol)
	newShort := phpShortName(input.NewSymbol)
	if oldShort == "" || oldShort == newShort || !hasPlainUseImport(document, input.OldSymbol, oldShort) {
		return nil
	}

	skippedRanges := document.NodesByKind("namespace_use_declaration", "namespace_definition")
	var result []replacements.Replacement
	for _, node := range document.NodesByKind("name", "qualified_name") {
		if node.Text != oldShort || isInsidePHPRange(node, skippedRanges) {
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
		keyword == "catch"
}

func nextNonSpace(source []byte, index int) int {
	for index < len(source) {
		if !isPHPWhitespace(source[index]) {
			return index
		}
		index++
	}
	return -1
}

func previousNonSpace(source []byte, index int) int {
	for index >= 0 {
		if !isPHPWhitespace(source[index]) {
			return index
		}
		index--
	}
	return -1
}

func previousWord(source []byte, start int) string {
	index := previousNonSpace(source, start-1)
	if index < 0 || !isPHPIdentifierByte(source[index]) {
		return ""
	}

	end := index + 1
	for index >= 0 && isPHPIdentifierByte(source[index]) {
		index--
	}
	return string(source[index+1 : end])
}

func isPHPWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func isPHPIdentifierByte(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func phpShortName(symbol string) string {
	index := strings.LastIndex(symbol, "\\")
	if index < 0 {
		return symbol
	}
	return symbol[index+1:]
}

func isInsidePHPRange(node treesitter.Node, ranges []treesitter.Node) bool {
	for _, candidate := range ranges {
		if node.StartByte >= candidate.StartByte && node.EndByte <= candidate.EndByte {
			return true
		}
	}
	return false
}
