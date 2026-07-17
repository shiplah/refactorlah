//go:build cgo

package rules

import (
	"sort"
	"strings"

	"github.com/shiplah/refactorlah/internal/adapters/php/names"
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"
	"github.com/shiplah/refactorlah/internal/replacements"
)

const SameNamespaceSymbolImportRuleName = "php.SameNamespaceSymbolImportRule"

type SameNamespaceSymbolImportInput struct {
	File     string
	Source   []byte
	Mappings []SymbolMappingReference
}

type SameNamespaceSymbolImportRule struct{}

func (r SameNamespaceSymbolImportRule) Collect(document *treesitter.Document, input SameNamespaceSymbolImportInput) []replacements.Replacement {
	namespace := declaredNamespace(document)
	if namespace == "" {
		return nil
	}

	skippedRanges := document.NodesByKind(
		"namespace_use_declaration",
		"namespace_definition",
		"const_declaration",
		"function_definition",
		"class_constant_access_expression",
	)
	existingImports := existingFunctionAndConstantImports(document)
	plannedImports := map[string]symbolImport{}

	for _, mapping := range input.Mappings {
		if mapping.Kind != "constant" && mapping.Kind != "function" {
			continue
		}
		if names.Namespace(mapping.OldSymbol) != namespace || names.Namespace(mapping.NewSymbol) == namespace {
			continue
		}

		shortName := names.Short(mapping.OldSymbol)
		if shortName == "" || existingImports[mapping.Kind][shortName] != "" {
			continue
		}

		for _, node := range document.NodesByKind("name", "qualified_name") {
			if node.Text != shortName || treesitter.NodeInsideAnyRange(node, skippedRanges) {
				continue
			}
			if isQualifiedNameSegment(input.Source, node.StartByte, node.EndByte) {
				continue
			}
			if mapping.Kind == "function" && !isUnqualifiedFunctionCall(input.Source, node.StartByte, node.EndByte) {
				continue
			}
			if mapping.Kind == "constant" && !isUnqualifiedConstantReference(input.Source, node.Text, node.StartByte, node.EndByte) {
				continue
			}

			plannedImports[mapping.Kind+"\\"+shortName] = symbolImport{
				kind:   mapping.Kind,
				symbol: mapping.NewSymbol,
			}
		}
	}

	if len(plannedImports) == 0 {
		return nil
	}
	if insertion, ok := sameNamespaceSymbolImportInsertion(document, plannedImports, input.File); ok {
		return []replacements.Replacement{insertion}
	}
	return nil
}

type symbolImport struct {
	kind   string
	symbol string
}

func existingFunctionAndConstantImports(document *treesitter.Document) map[string]map[string]string {
	imports := map[string]map[string]string{
		"constant": {},
		"function": {},
	}
	for _, node := range document.NodesByKind("namespace_use_declaration") {
		kind, symbol, ok := plainFunctionOrConstantImport(node.Text)
		if !ok {
			continue
		}
		imports[kind][names.Short(symbol)] = symbol
	}
	return imports
}

func plainFunctionOrConstantImport(useStatement string) (string, string, bool) {
	text := strings.TrimSpace(useStatement)
	if !strings.HasPrefix(text, "use ") || !strings.HasSuffix(text, ";") {
		return "", "", false
	}
	if strings.Contains(strings.ToLower(text), " as ") || strings.Contains(text, "{") || strings.Contains(text, ",") {
		return "", "", false
	}

	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "use "), ";"))
	lower := strings.ToLower(body)
	switch {
	case strings.HasPrefix(lower, "const "):
		return "constant", strings.TrimSpace(body[len("const "):]), true
	case strings.HasPrefix(lower, "function "):
		return "function", strings.TrimSpace(body[len("function "):]), true
	default:
		return "", "", false
	}
}

func isUnqualifiedFunctionCall(source []byte, start int, end int) bool {
	if isMemberAccessName(source, start) {
		return false
	}
	next := nextNonSpace(source, end)
	return next >= 0 && source[next] == '('
}

func isUnqualifiedConstantReference(source []byte, name string, start int, end int) bool {
	if isPHPMagicConstant(name) {
		return false
	}
	if start > 0 && source[start-1] == '$' {
		return false
	}
	if isMemberAccessName(source, start) {
		return false
	}
	if isUnqualifiedFunctionCall(source, start, end) || isRightSideOfClassConstantAccess(source, start) {
		return false
	}

	return true
}

func isMemberAccessName(source []byte, start int) bool {
	previous := previousNonSpace(source, start-1)
	if previous < 1 {
		return false
	}

	if source[previous] == '>' && source[previous-1] == '-' {
		return true
	}
	return source[previous] == ':' && source[previous-1] == ':'
}

func sameNamespaceSymbolImportInsertion(document *treesitter.Document, imports map[string]symbolImport, file string) (replacements.Replacement, bool) {
	rendered := renderSymbolUseStatements(imports)
	useStatements := document.NodesByKind("namespace_use_declaration")
	if lastUse, ok := lastUseStatement(useStatements); ok {
		return replacements.Replacement{
			File:        file,
			Start:       lastUse.EndByte,
			End:         lastUse.EndByte,
			Replacement: "\n" + rendered,
			Reason:      "php-same-namespace-symbol-import",
			Rule:        SameNamespaceSymbolImportRuleName,
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
			File:        file,
			Start:       offset,
			End:         offset,
			Replacement: "\n\n" + rendered,
			Reason:      "php-same-namespace-symbol-import",
			Rule:        SameNamespaceSymbolImportRuleName,
			Adapter:     "php",
		}, true
	}

	return replacements.Replacement{}, false
}

func lastUseStatement(useStatements []treesitter.Node) (treesitter.Node, bool) {
	if len(useStatements) == 0 {
		return treesitter.Node{}, false
	}
	return useStatements[len(useStatements)-1], true
}

func renderSymbolUseStatements(imports map[string]symbolImport) string {
	symbols := make([]symbolImport, 0, len(imports))
	for _, imported := range imports {
		symbols = append(symbols, imported)
	}
	sort.Slice(symbols, func(left int, right int) bool {
		leftOrder := symbolImportKindOrder(symbols[left].kind)
		rightOrder := symbolImportKindOrder(symbols[right].kind)
		if leftOrder != rightOrder {
			return leftOrder < rightOrder
		}
		return symbols[left].symbol < symbols[right].symbol
	})

	lines := make([]string, 0, len(symbols))
	for _, imported := range symbols {
		switch imported.kind {
		case "constant":
			lines = append(lines, "use const "+imported.symbol+";")
		case "function":
			lines = append(lines, "use function "+imported.symbol+";")
		}
	}
	return strings.Join(lines, "\n")
}

func symbolImportKindOrder(kind string) int {
	switch kind {
	case "constant":
		return 0
	case "function":
		return 1
	default:
		return 2
	}
}
