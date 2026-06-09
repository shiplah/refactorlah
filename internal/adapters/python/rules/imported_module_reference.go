//go:build cgo

package rules

import (
	"strings"

	"github.com/shiplah/refactorlah/internal/adapters/python/syntax"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/replacements"
)

const ImportedModuleReferenceRuleName = "python.ImportedModuleReferenceRule"

type ImportedModuleReferenceInput struct {
	File      string
	Package   string
	Source    []byte
	OldModule string
	NewModule string
}

type ImportedModuleReferenceRule struct{}

func (r ImportedModuleReferenceRule) Collect(document *treesitter.Document, input ImportedModuleReferenceInput) []replacements.Replacement {
	oldLeaf := syntax.Leaf(input.OldModule)
	newLeaf := syntax.Leaf(input.NewModule)
	oldParent := syntax.Parent(input.OldModule)
	if oldLeaf == "" || oldLeaf == newLeaf || oldParent == "" {
		return nil
	}
	if !importsVisibleLeaf(document, input.Package, oldParent, oldLeaf) {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("identifier") {
		if node.Text != oldLeaf || !isFollowedByDot(input.Source, node.EndByte) {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte,
			End:         node.EndByte,
			Replacement: newLeaf,
			Reason:      "python-imported-module-reference",
			Rule:        ImportedModuleReferenceRuleName,
			Adapter:     "python",
		})
	}

	return result
}

func importsVisibleLeaf(document *treesitter.Document, packageName string, parent string, leaf string) bool {
	for _, node := range document.NodesByKind("import_from_statement") {
		resolvedParent, importClause, ok := fromImportParts(node.Text, packageName)
		if !ok || resolvedParent != parent {
			continue
		}
		if importsUnaliasedName(importClause, leaf) {
			return true
		}
	}

	return false
}

func fromImportParts(text string, packageName string) (string, string, bool) {
	trimmedStart := strings.TrimLeft(text, " \t")
	if !strings.HasPrefix(trimmedStart, "from ") {
		return "", "", false
	}

	fromIndex := strings.Index(text, "from ")
	moduleStart := fromIndex + len("from ")
	importMarker := " import "
	importIndex := strings.Index(text[moduleStart:], importMarker)
	if importIndex < 0 {
		return "", "", false
	}

	moduleEnd := moduleStart + importIndex
	moduleText := strings.TrimSpace(text[moduleStart:moduleEnd])
	importClause := text[moduleEnd+len(importMarker):]
	if strings.HasPrefix(moduleText, ".") {
		dotCount := 0
		for dotCount < len(moduleText) && moduleText[dotCount] == '.' {
			dotCount++
		}
		moduleTail := moduleText[dotCount:]
		resolved, ok := syntax.ResolveRelativeModule(packageName, dotCount, moduleTail)
		return resolved, importClause, ok
	}

	return moduleText, importClause, true
}

func importsUnaliasedName(importClause string, name string) bool {
	for _, item := range strings.Split(importClause, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		parts := strings.Fields(item)
		if len(parts) == 1 && parts[0] == name {
			return true
		}
	}

	return false
}

func isFollowedByDot(source []byte, offset int) bool {
	return offset >= 0 && offset < len(source) && source[offset] == '.'
}
