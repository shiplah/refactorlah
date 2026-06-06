//go:build cgo

package rules

import (
	"unicode"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const QualifiedModuleReferenceRuleName = "python.QualifiedModuleReferenceRule"

type QualifiedModuleReferenceInput struct {
	File      string
	OldModule string
	NewModule string
}

type QualifiedModuleReferenceRule struct{}

func (r QualifiedModuleReferenceRule) Collect(document *treesitter.Document, input QualifiedModuleReferenceInput) []replacements.Replacement {
	if input.OldModule == "" || input.OldModule == input.NewModule {
		return nil
	}

	skippedRanges := document.NodesByKind("import_statement", "import_from_statement")
	var result []replacements.Replacement
	seen := map[qualifiedReplacementKey]bool{}
	for _, node := range document.NodesByKind("attribute", "dotted_name") {
		if treesitter.NodeInsideAnyRange(node, skippedRanges) {
			continue
		}

		start := findPythonQualifiedModuleOccurrence(node.Text, input.OldModule)
		if start < 0 {
			continue
		}

		startOffset := node.StartByte + start
		endOffset := startOffset + len(input.OldModule)
		key := qualifiedReplacementKey{
			start:       startOffset,
			end:         endOffset,
			replacement: input.NewModule,
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       startOffset,
			End:         endOffset,
			Replacement: input.NewModule,
			Reason:      "python-qualified-module",
			Rule:        QualifiedModuleReferenceRuleName,
			Adapter:     "python",
		})
	}

	return result
}

type qualifiedReplacementKey struct {
	start       int
	end         int
	replacement string
}

func findPythonQualifiedModuleOccurrence(text string, module string) int {
	offset := 0
	for {
		index := findSubstring(text[offset:], module)
		if index < 0 {
			return -1
		}

		start := offset + index
		end := start + len(module)
		if isQualifiedModuleStartBoundary(text, start-1) && isQualifiedModuleEndBoundary(text, end) {
			return start
		}

		offset = end
	}
}

func findSubstring(text string, needle string) int {
	for index := 0; index+len(needle) <= len(text); index++ {
		if text[index:index+len(needle)] == needle {
			return index
		}
	}
	return -1
}

func isQualifiedModuleStartBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	character := rune(text[index])
	return character != '.' && character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character)
}

func isQualifiedModuleEndBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	character := rune(text[index])
	return character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character)
}
