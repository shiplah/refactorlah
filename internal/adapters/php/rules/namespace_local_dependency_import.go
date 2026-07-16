//go:build cgo

package rules

import (
	"path"
	"sort"
	"strings"

	"github.com/shiplah/refactorlah/internal/adapters/php/names"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/replacements"
)

const NamespaceLocalDependencyImportRuleName = "php.NamespaceLocalDependencyImportRule"

type NamespaceLocalDependencyImportInput struct {
	File         string
	Source       []byte
	OldNamespace string
	NewNamespace string
	Mappings     []SymbolMappingReference
}

type NamespaceLocalDependencyImportRule struct{}

func (r NamespaceLocalDependencyImportRule) Collect(document *treesitter.Document, input NamespaceLocalDependencyImportInput) []replacements.Replacement {
	if input.OldNamespace == "" || input.NewNamespace == "" || input.OldNamespace == input.NewNamespace {
		return nil
	}
	if declaredNamespace(document) != input.OldNamespace {
		return nil
	}

	declaredNames := declaredClassLikeNames(document)
	existingImports := existingNormalImports(document)
	plannedImports := map[string]string{}
	var result []replacements.Replacement
	skippedRanges := document.NodesByKind("namespace_use_declaration", "namespace_definition")

	for _, node := range document.NodesByKind("name", "qualified_name") {
		if node.Text == "" || strings.Contains(node.Text, "\\") || declaredNames[node.Text] || treesitter.NodeInsideAnyRange(node, skippedRanges) {
			continue
		}
		if isPHPMagicConstant(node.Text) {
			continue
		}
		if !isSafeShortClassReference(input.Source, node.StartByte, node.EndByte) || !isLikelyClassName(node.Text) {
			continue
		}
		desiredSymbol := desiredNamespaceLocalSymbol(input, node.Text)
		if names.Namespace(desiredSymbol) == input.NewNamespace {
			continue
		}
		if existingImports[node.Text] != "" {
			continue
		}
		if plannedImports[node.Text] != "" && plannedImports[node.Text] != desiredSymbol {
			continue
		}

		plannedImports[node.Text] = desiredSymbol
	}

	if len(plannedImports) == 0 {
		return result
	}
	if insertion, ok := namespaceLocalImportInsertion(document, plannedImports, input); ok {
		result = append(result, insertion)
	}

	return result
}

func declaredClassLikeNames(document *treesitter.Document) map[string]bool {
	names := map[string]bool{}
	for _, node := range document.NodesByKind("class_declaration", "interface_declaration", "trait_declaration", "enum_declaration") {
		for _, word := range strings.Fields(node.Text) {
			candidate := strings.Trim(word, "{")
			if candidate == "class" || candidate == "interface" || candidate == "trait" || candidate == "enum" {
				continue
			}
			if isLikelyClassName(candidate) {
				names[candidate] = true
				break
			}
		}
	}
	return names
}

func desiredNamespaceLocalSymbol(input NamespaceLocalDependencyImportInput, shortName string) string {
	oldSymbol := input.OldNamespace + "\\" + shortName
	for _, mapping := range input.Mappings {
		if mapping.OldSymbol == oldSymbol {
			return mapping.NewSymbol
		}
	}
	return oldSymbol
}

func namespaceLocalImportInsertion(document *treesitter.Document, imports map[string]string, input NamespaceLocalDependencyImportInput) (replacements.Replacement, bool) {
	symbols := make([]string, 0, len(imports))
	for _, symbol := range imports {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	rendered := renderUseStatements(symbols)
	useStatements := document.NodesByKind("namespace_use_declaration")
	if lastUse, ok := lastPreservedClassUseStatement(useStatements, input); ok {
		return replacements.Replacement{
			File:        input.File,
			Start:       lastUse.EndByte,
			End:         lastUse.EndByte,
			Replacement: "\n" + rendered,
			Reason:      "php-namespace-local-import",
			Rule:        NamespaceLocalDependencyImportRuleName,
			Adapter:     "php",
		}, true
	}
	if firstUse, ok := firstUseStatement(useStatements); ok && !isClassUseStatement(firstUse.Text) {
		return replacements.Replacement{
			File:        input.File,
			Start:       firstUse.StartByte,
			End:         firstUse.StartByte,
			Replacement: rendered + "\n\n",
			Reason:      "php-namespace-local-import",
			Rule:        NamespaceLocalDependencyImportRuleName,
			Adapter:     "php",
		}, true
	}

	for _, namespace := range document.NodesByKind("namespace_definition") {
		semicolon := strings.Index(namespace.Text, ";")
		if semicolon < 0 {
			return replacements.Replacement{}, false
		}

		offset := namespace.StartByte + semicolon + 1
		return replacements.Replacement{
			File:        input.File,
			Start:       offset,
			End:         offset,
			Replacement: "\n\n" + rendered,
			Reason:      "php-namespace-local-import",
			Rule:        NamespaceLocalDependencyImportRuleName,
			Adapter:     "php",
		}, true
	}

	return replacements.Replacement{}, false
}

func lastPreservedClassUseStatement(useStatements []treesitter.Node, input NamespaceLocalDependencyImportInput) (treesitter.Node, bool) {
	for index := len(useStatements) - 1; index >= 0; index-- {
		if !isClassUseStatement(useStatements[index].Text) {
			continue
		}
		importedSymbol, ok := plainImportedSymbol(useStatements[index].Text)
		if ok && importBecomesSameNamespace(importedSymbol, input.NewNamespace, input.Mappings) {
			continue
		}

		return useStatements[index], true
	}

	return treesitter.Node{}, false
}

func renderUseStatements(symbols []string) string {
	lines := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		lines = append(lines, "use "+symbol+";")
	}
	return strings.Join(lines, "\n")
}

func isLikelyClassName(name string) bool {
	if name == "" || strings.Contains(name, "\\") || path.Ext(name) != "" || isPHPBuiltinType(strings.ToLower(name)) {
		return false
	}

	first := name[0]
	return first == '_' || first >= 'A' && first <= 'Z'
}

func isPHPBuiltinType(name string) bool {
	switch name {
	case "array", "bool", "callable", "false", "float", "int", "iterable", "mixed", "never", "null", "object", "parent", "self", "static", "string", "true", "void":
		return true
	default:
		return false
	}
}

func isPHPMagicConstant(name string) bool {
	switch strings.ToUpper(name) {
	case "__CLASS__",
		"__DIR__",
		"__FILE__",
		"__FUNCTION__",
		"__LINE__",
		"__METHOD__",
		"__NAMESPACE__",
		"__PROPERTY__",
		"__TRAIT__":
		return true
	default:
		return false
	}
}
