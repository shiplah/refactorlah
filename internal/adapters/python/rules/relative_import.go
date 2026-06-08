//go:build cgo

package rules

import (
	"regexp"

	"github.com/NickSdot/refactorlah/internal/adapters/python/syntax"
	"github.com/NickSdot/refactorlah/internal/parsing/treesitter"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

const RelativeImportRuleName = "python.RelativeImportRule"

type RelativeImportInput struct {
	File      string
	Package   string
	OldModule string
	NewModule string
}

type RelativeImportRule struct{}

var relativeFromImportPattern = regexp.MustCompile(`from[ \t]+(\.+)([A-Za-z_][\w.]*)?([ \t]+import[ \t]+)([^\n#]+)`)

func (r RelativeImportRule) Collect(document *treesitter.Document, input RelativeImportInput) []replacements.Replacement {
	if input.Package == "" || input.OldModule == "" || input.OldModule == input.NewModule {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("import_from_statement") {
		match := relativeFromImportPattern.FindStringSubmatchIndex(node.Text)
		if match == nil {
			continue
		}

		dotsStart, dotsEnd := match[2], match[3]
		moduleTailStart, moduleTailEnd := match[4], match[5]
		importClauseStart, importClauseEnd := match[8], match[9]

		moduleTail := ""
		if moduleTailStart >= 0 {
			moduleTail = node.Text[moduleTailStart:moduleTailEnd]
		}

		resolvedModule, ok := syntax.ResolveRelativeModule(input.Package, dotsEnd-dotsStart, moduleTail)
		if !ok {
			continue
		}

		if moduleTail != "" {
			if resolvedModule != input.OldModule {
				continue
			}
			result = append(result, replacements.Replacement{
				File:        input.File,
				Start:       node.StartByte + dotsStart,
				End:         node.StartByte + moduleTailEnd,
				Replacement: input.NewModule,
				Reason:      "python-relative-from-import",
				Rule:        RelativeImportRuleName,
				Adapter:     "python",
			})
			continue
		}

		oldParent := syntax.Parent(input.OldModule)
		newParent := syntax.Parent(input.NewModule)
		oldLeaf := syntax.Leaf(input.OldModule)
		newLeaf := syntax.Leaf(input.NewModule)
		if resolvedModule != oldParent || newParent == "" {
			continue
		}

		importClause := node.Text[importClauseStart:importClauseEnd]
		leafStart := findPythonModuleOccurrence(importClause, oldLeaf)
		if leafStart < 0 {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + dotsStart,
			End:         node.StartByte + dotsEnd,
			Replacement: newParent,
			Reason:      "python-relative-from-import",
			Rule:        RelativeImportRuleName,
			Adapter:     "python",
		})

		if oldLeaf != newLeaf {
			result = append(result, replacements.Replacement{
				File:        input.File,
				Start:       node.StartByte + importClauseStart + leafStart,
				End:         node.StartByte + importClauseStart + leafStart + len(oldLeaf),
				Replacement: newLeaf,
				Reason:      "python-relative-from-import-name",
				Rule:        RelativeImportRuleName,
				Adapter:     "python",
			})
		}
	}

	return result
}
