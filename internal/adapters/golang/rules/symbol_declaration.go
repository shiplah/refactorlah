package rules

import (
	"go/ast"
	"go/parser"
	"go/token"

	adapterproto "refactorlah/internal/adapters/contract"
)

const SymbolDeclarationRuleName = "go.SymbolDeclarationRule"

type SymbolDeclarationMapping struct {
	File      string
	OldSymbol string
	NewSymbol string
	Kind      string
}

type SymbolDeclarationInput struct {
	File     string
	Mappings []SymbolDeclarationMapping
}

type SymbolDeclarationRule struct{}

func (r SymbolDeclarationRule) Collect(source []byte, input SymbolDeclarationInput) ([]adapterproto.Replacement, error) {
	if len(input.Mappings) == 0 {
		return nil, nil
	}

	mappings := map[string]SymbolDeclarationMapping{}
	for _, mapping := range input.Mappings {
		if mapping.File != input.File || mapping.OldSymbol == "" || mapping.OldSymbol == mapping.NewSymbol {
			continue
		}
		mappings[mapping.OldSymbol] = mapping
	}
	if len(mappings) == 0 {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, input.File, source, 0)
	if err != nil {
		return nil, err
	}

	var replacements []adapterproto.Replacement
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			replacements = append(replacements, declarationReplacement(fileSet, input.File, typed.Name, mappings)...)
		case *ast.GenDecl:
			for _, spec := range typed.Specs {
				switch specTyped := spec.(type) {
				case *ast.TypeSpec:
					replacements = append(replacements, declarationReplacement(fileSet, input.File, specTyped.Name, mappings)...)
				case *ast.ValueSpec:
					for _, name := range specTyped.Names {
						replacements = append(replacements, declarationReplacement(fileSet, input.File, name, mappings)...)
					}
				}
			}
		}
	}

	return replacements, nil
}

func declarationReplacement(fileSet *token.FileSet, filePath string, ident *ast.Ident, mappings map[string]SymbolDeclarationMapping) []adapterproto.Replacement {
	if ident == nil {
		return nil
	}
	mapping, ok := mappings[ident.Name]
	if !ok {
		return nil
	}

	tokenFile := fileSet.File(ident.Pos())
	return []adapterproto.Replacement{{
		File:        filePath,
		Start:       tokenFile.Offset(ident.Pos()),
		End:         tokenFile.Offset(ident.End()),
		Replacement: mapping.NewSymbol,
		Reason:      "go-symbol-declaration",
		Rule:        SymbolDeclarationRuleName,
		Adapter:     "go",
	}}
}
