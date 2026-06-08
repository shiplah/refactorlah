//go:build cgo

package rules

import (
	"strings"
	"unicode"

	"github.com/NickSdot/refactorlah/internal/adapters/python/syntax"
	"github.com/NickSdot/refactorlah/internal/parsing/treesitter"
	"github.com/NickSdot/refactorlah/internal/replacements"
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
		if start >= 0 {
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

		if node.Kind == "import_from_statement" {
			result = append(result, collectFromParentImportReplacements(node, input)...)
		}
	}

	return result
}

func collectFromParentImportReplacements(node treesitter.Node, input ImportStatementInput) []replacements.Replacement {
	oldParent := syntax.Parent(input.OldModule)
	newParent := syntax.Parent(input.NewModule)
	oldLeaf := syntax.Leaf(input.OldModule)
	newLeaf := syntax.Leaf(input.NewModule)
	if oldParent == "" {
		return nil
	}

	prefix := "from "
	if !strings.HasPrefix(strings.TrimLeft(node.Text, " \t"), prefix) {
		return nil
	}
	fromIndex := strings.Index(node.Text, prefix)
	parentStart := fromIndex + len(prefix)
	importMarker := " import "
	importIndex := strings.Index(node.Text[parentStart:], importMarker)
	if importIndex < 0 {
		return nil
	}
	parentEnd := parentStart + importIndex
	parent := strings.TrimSpace(node.Text[parentStart:parentEnd])
	if parent != oldParent {
		return nil
	}

	importClauseStart := parentEnd + len(importMarker)
	importClause := node.Text[importClauseStart:]
	leafStarts := importedNameOffsets(importClause, oldLeaf)
	if len(leafStarts) == 0 {
		return nil
	}

	var result []replacements.Replacement
	if newParent != "" && oldParent != newParent {
		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + parentStart,
			End:         node.StartByte + parentStart + len(oldParent),
			Replacement: newParent,
			Reason:      "python-from-import",
			Rule:        ImportStatementRuleName,
			Adapter:     "python",
		})
	}

	if oldLeaf != newLeaf {
		for _, leafStart := range leafStarts {
			result = append(result, replacements.Replacement{
				File:        input.File,
				Start:       node.StartByte + importClauseStart + leafStart,
				End:         node.StartByte + importClauseStart + leafStart + len(oldLeaf),
				Replacement: newLeaf,
				Reason:      "python-from-import-name",
				Rule:        ImportStatementRuleName,
				Adapter:     "python",
			})
		}
	}

	return result
}

func importedNameOffsets(importClause string, name string) []int {
	var offsets []int
	offset := 0
	for {
		index := strings.Index(importClause[offset:], name)
		if index < 0 {
			return offsets
		}

		start := offset + index
		end := start + len(name)
		if isPythonModuleBoundary(importClause, start-1) && isPythonModuleBoundary(importClause, end) {
			offsets = append(offsets, start)
		}

		offset = end
	}
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
