//go:build cgo

package rules

import (
	"strings"
	"unicode"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const UseStatementRuleName = "php.UseStatementRule"

type UseStatementInput struct {
	File      string
	OldSymbol string
	NewSymbol string
}

type UseStatementRule struct{}

func (r UseStatementRule) Collect(document *treesitter.Document, input UseStatementInput) []replacements.Replacement {
	if input.OldSymbol == "" || input.OldSymbol == input.NewSymbol {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		start := findPHPNameOccurrence(node.Text, input.OldSymbol)
		if start < 0 {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + start,
			End:         node.StartByte + start + len(input.OldSymbol),
			Replacement: input.NewSymbol,
			Reason:      "php-use-statement",
			Rule:        UseStatementRuleName,
			Adapter:     "php",
		})
	}

	return result
}

func findPHPNameOccurrence(text string, name string) int {
	offset := 0
	for {
		index := strings.Index(text[offset:], name)
		if index < 0 {
			return -1
		}

		start := offset + index
		end := start + len(name)
		if isPHPNameBoundary(text, start-1) && isPHPNameBoundary(text, end) {
			return start
		}

		offset = end
	}
}

func isPHPNameBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	character := rune(text[index])
	return character != '\\' && character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character)
}
