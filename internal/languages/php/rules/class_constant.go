//go:build cgo

package rules

import (
	"strings"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const ClassConstantRuleName = "php.ClassConstantRule"

type ClassConstantInput struct {
	File      string
	OldSymbol string
	NewSymbol string
}

type ClassConstantRule struct{}

func (r ClassConstantRule) Collect(document *treesitter.Document, input ClassConstantInput) []replacements.Replacement {
	oldSymbol := strings.TrimPrefix(input.OldSymbol, "\\")
	newSymbol := strings.TrimPrefix(input.NewSymbol, "\\")
	if oldSymbol == "" || oldSymbol == newSymbol {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("class_constant_access_expression") {
		nameStart, nameEnd, prefix, ok := classNameBeforeClassConstant(node.Text)
		if !ok {
			continue
		}
		if strings.TrimPrefix(node.Text[nameStart:nameEnd], "\\") != oldSymbol {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + nameStart,
			End:         node.StartByte + nameEnd,
			Replacement: prefix + newSymbol,
			Reason:      "php-class-constant",
			Rule:        ClassConstantRuleName,
			Adapter:     "php",
		})
	}

	return result
}

func classNameBeforeClassConstant(text string) (int, int, string, bool) {
	classIndex := strings.Index(text, "::class")
	if classIndex < 0 {
		return 0, 0, "", false
	}

	nameStart := 0
	for nameStart < classIndex && isSpaceByte(text[nameStart]) {
		nameStart++
	}
	nameEnd := classIndex
	for nameEnd > nameStart && isSpaceByte(text[nameEnd-1]) {
		nameEnd--
	}
	if nameEnd <= nameStart {
		return 0, 0, "", false
	}

	prefix := ""
	if text[nameStart] == '\\' {
		prefix = "\\"
	}

	return nameStart, nameEnd, prefix, true
}

func isSpaceByte(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}
