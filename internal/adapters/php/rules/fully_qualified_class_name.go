//go:build cgo

package rules

import (
	"strings"

	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const FullyQualifiedClassNameRuleName = "php.FullyQualifiedClassNameRule"

type FullyQualifiedClassNameInput struct {
	File      string
	OldSymbol string
	NewSymbol string
}

type FullyQualifiedClassNameRule struct{}

func (r FullyQualifiedClassNameRule) Collect(document *treesitter.Document, input FullyQualifiedClassNameInput) []replacements.Replacement {
	oldSymbol := strings.TrimPrefix(input.OldSymbol, "\\")
	newSymbol := strings.TrimPrefix(input.NewSymbol, "\\")
	if oldSymbol == "" || oldSymbol == newSymbol {
		return nil
	}

	skippedRanges := document.NodesByKind("namespace_definition", "namespace_use_declaration", "class_constant_access_expression")
	var result []replacements.Replacement
	for _, node := range document.NodesByKind("qualified_name") {
		if treesitter.NodeInsideAnyRange(node, skippedRanges) {
			continue
		}

		prefix := ""
		candidate := node.Text
		if strings.HasPrefix(candidate, "\\") {
			prefix = "\\"
			candidate = strings.TrimPrefix(candidate, "\\")
		}
		if candidate != oldSymbol {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte,
			End:         node.EndByte,
			Replacement: prefix + newSymbol,
			Reason:      "php-fully-qualified-class-name",
			Rule:        FullyQualifiedClassNameRuleName,
			Adapter:     "php",
		})
	}

	return result
}
