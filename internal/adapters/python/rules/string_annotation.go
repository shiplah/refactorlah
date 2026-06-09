//go:build cgo

package rules

import (
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/replacements"
)

const StringAnnotationRuleName = "python.StringAnnotationRule"

type StringAnnotationInput struct {
	File      string
	OldModule string
	NewModule string
}

type StringAnnotationRule struct{}

func (r StringAnnotationRule) Collect(document *treesitter.Document, input StringAnnotationInput) []replacements.Replacement {
	if input.OldModule == "" || input.OldModule == input.NewModule {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("string") {
		if !isAnnotationString(node) {
			continue
		}

		offset := 0
		for {
			start := findPythonQualifiedModuleOccurrence(node.Text[offset:], input.OldModule)
			if start < 0 {
				break
			}

			absoluteStart := node.StartByte + offset + start
			result = append(result, replacements.Replacement{
				File:        input.File,
				Start:       absoluteStart,
				End:         absoluteStart + len(input.OldModule),
				Replacement: input.NewModule,
				Reason:      "python-string-annotation",
				Rule:        StringAnnotationRuleName,
				Adapter:     "python",
			})
			offset += start + len(input.OldModule)
		}
	}

	return result
}

func isAnnotationString(node treesitter.Node) bool {
	return node.ParentKind() == "type" && isPlainAnnotationStringLiteral(node.Text)
}

func isPlainAnnotationStringLiteral(text string) bool {
	for _, character := range text {
		switch character {
		case '\'', '"':
			return true
		case 'b', 'B', 'f', 'F':
			return false
		case 'r', 'R', 'u', 'U':
			continue
		default:
			return false
		}
	}

	return false
}
