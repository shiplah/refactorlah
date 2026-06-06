package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	adapterproto "refactorlah/internal/adapters"
)

const ImportedSymbolReferenceRuleName = "go.ImportedSymbolReferenceRule"

type ImportedSymbolReferenceMapping struct {
	OldImport  string
	NewImport  string
	OldPackage string
	NewPackage string
	OldSymbol  string
	NewSymbol  string
}

type ImportedSymbolReferenceInput struct {
	File     string
	Mappings []ImportedSymbolReferenceMapping
}

type ImportedSymbolReferenceRule struct{}

func (r ImportedSymbolReferenceRule) Collect(source []byte, input ImportedSymbolReferenceInput) ([]adapterproto.Replacement, error) {
	if len(input.Mappings) == 0 {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, input.File, source, 0)
	if err != nil {
		return nil, err
	}

	importsByMapping := importedSymbolSelectors(file, input.Mappings)
	if len(importsByMapping) == 0 {
		return nil, nil
	}

	var replacements []adapterproto.Replacement
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		qualifier, ok := selector.X.(*ast.Ident)
		if !ok {
			return true
		}

		for index, names := range importsByMapping {
			if !names[qualifier.Name] {
				continue
			}
			mapping := input.Mappings[index]
			if selector.Sel == nil || selector.Sel.Name != mapping.OldSymbol || mapping.OldSymbol == mapping.NewSymbol {
				continue
			}

			tokenFile := fileSet.File(selector.Sel.Pos())
			replacements = append(replacements, adapterproto.Replacement{
				File:        input.File,
				Start:       tokenFile.Offset(selector.Sel.Pos()),
				End:         tokenFile.Offset(selector.Sel.End()),
				Replacement: mapping.NewSymbol,
				Reason:      "go-imported-symbol-reference",
				Rule:        ImportedSymbolReferenceRuleName,
				Adapter:     "go",
			})
		}

		return true
	})

	return replacements, nil
}

func importedSymbolSelectors(file *ast.File, mappings []ImportedSymbolReferenceMapping) map[int]map[string]bool {
	result := map[int]map[string]bool{}
	for index, mapping := range mappings {
		if mapping.OldImport == "" || mapping.OldSymbol == "" || mapping.OldSymbol == mapping.NewSymbol {
			continue
		}

		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil || importPath != mapping.OldImport {
				continue
			}

			name := mapping.OldPackage
			if importSpec.Name != nil {
				if importSpec.Name.Name == "." || importSpec.Name.Name == "_" {
					continue
				}
				name = importSpec.Name.Name
			}
			if name == "" {
				continue
			}
			if result[index] == nil {
				result[index] = map[string]bool{}
			}
			result[index][name] = true
		}
	}
	return result
}
