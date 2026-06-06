//go:build cgo

package rules

import (
	"strings"
	"unicode"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const ImportStatementRuleName = "python.ImportStatementRule"

type ImportStatementInput struct {
	File      string
	OldModule string
	NewModule string
}

type ImportStatementRule struct{}

func (r ImportStatementRule) Collect(document *treesitter.Document, input ImportStatementInput) []replacements.Replacement {
	if input.OldModule == "" || input.OldModule == input.NewModule {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("import_statement", "import_from_statement") {
		start := findPythonModuleOccurrence(node.Text, input.OldModule)
		if start < 0 {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + start,
			End:         node.StartByte + start + len(input.OldModule),
			Replacement: input.NewModule,
			Reason:      "python-import-statement",
			Rule:        ImportStatementRuleName,
			Adapter:     "python",
		})
	}

	return result
}

func findPythonModuleOccurrence(text string, module string) int {
	offset := 0
	for {
		index := strings.Index(text[offset:], module)
		if index < 0 {
			return -1
		}

		start := offset + index
		end := start + len(module)
		if isPythonModuleBoundary(text, start-1) && isPythonModuleBoundary(text, end) {
			return start
		}

		offset = end
	}
}

func isPythonModuleBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	character := rune(text[index])
	return character != '.' && character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character)
}
