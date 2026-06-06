package rules

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"sort"

	adapterproto "refactorlah/internal/adapters"
)

const LocalSymbolReferenceRuleName = "go.LocalSymbolReferenceRule"

type GoSourceFile struct {
	File   string
	Source []byte
}

type LocalSymbolReferenceMapping struct {
	File      string
	OldSymbol string
	NewSymbol string
}

type LocalSymbolReferenceInput struct {
	PackageName string
	Files       []GoSourceFile
	Mappings    []LocalSymbolReferenceMapping
}

type LocalSymbolReferenceRule struct{}

func (r LocalSymbolReferenceRule) Collect(input LocalSymbolReferenceInput) ([]adapterproto.Replacement, error) {
	if input.PackageName == "" || len(input.Files) == 0 || len(input.Mappings) == 0 {
		return nil, nil
	}

	parsedPackage, err := parseGoPackage(input.PackageName, input.Files)
	if err != nil {
		return nil, err
	}
	if len(parsedPackage.files) == 0 {
		return nil, nil
	}

	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{},
	}
	config := types.Config{
		Importer: importer.Default(),
		Error:    func(error) {},
	}
	_, _ = config.Check(input.PackageName, parsedPackage.fileSet, parsedPackage.files, info)

	objectsByMapping := topLevelObjectsByMapping(parsedPackage, info, input.Mappings)
	if len(objectsByMapping) == 0 {
		return nil, nil
	}

	var replacements []adapterproto.Replacement
	for ident, object := range info.Uses {
		for index, target := range objectsByMapping {
			if object != target {
				continue
			}
			mapping := input.Mappings[index]
			if ident.Name != mapping.OldSymbol {
				continue
			}
			tokenFile := parsedPackage.fileSet.File(ident.Pos())
			replacements = append(replacements, adapterproto.Replacement{
				File:        parsedPackage.pathByTokenFile[tokenFile],
				Start:       tokenFile.Offset(ident.Pos()),
				End:         tokenFile.Offset(ident.End()),
				Replacement: mapping.NewSymbol,
				Reason:      "go-local-symbol-reference",
				Rule:        LocalSymbolReferenceRuleName,
				Adapter:     "go",
			})
		}
	}

	sort.Slice(replacements, func(left int, right int) bool {
		if replacements[left].File == replacements[right].File {
			return replacements[left].Start < replacements[right].Start
		}
		return replacements[left].File < replacements[right].File
	})

	return replacements, nil
}

type parsedGoPackage struct {
	fileSet         *token.FileSet
	files           []*ast.File
	pathByAST       map[*ast.File]string
	pathByTokenFile map[*token.File]string
}

func parseGoPackage(packageName string, files []GoSourceFile) (parsedGoPackage, error) {
	fileSet := token.NewFileSet()
	parsed := parsedGoPackage{
		fileSet:         fileSet,
		pathByAST:       map[*ast.File]string{},
		pathByTokenFile: map[*token.File]string{},
	}

	for _, sourceFile := range files {
		file, err := parser.ParseFile(fileSet, sourceFile.File, sourceFile.Source, 0)
		if err != nil {
			return parsedGoPackage{}, err
		}
		if file.Name == nil || file.Name.Name != packageName {
			continue
		}
		parsed.files = append(parsed.files, file)
		parsed.pathByAST[file] = sourceFile.File
		if tokenFile := fileSet.File(file.Package); tokenFile != nil {
			parsed.pathByTokenFile[tokenFile] = sourceFile.File
		}
		for _, declaration := range file.Decls {
			tokenFile := fileSet.File(declaration.Pos())
			if tokenFile != nil {
				parsed.pathByTokenFile[tokenFile] = sourceFile.File
			}
		}
	}

	return parsed, nil
}

func topLevelObjectsByMapping(parsedPackage parsedGoPackage, info *types.Info, mappings []LocalSymbolReferenceMapping) map[int]types.Object {
	objects := map[int]types.Object{}
	for index, mapping := range mappings {
		for _, file := range parsedPackage.files {
			if parsedPackage.pathByAST[file] != mapping.File {
				continue
			}
			ident := topLevelDeclarationIdent(file, mapping.OldSymbol)
			if ident == nil {
				continue
			}
			object := info.Defs[ident]
			if object != nil {
				objects[index] = object
			}
		}
	}
	return objects
}

func topLevelDeclarationIdent(file *ast.File, name string) *ast.Ident {
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if typed.Recv == nil && typed.Name != nil && typed.Name.Name == name {
				return typed.Name
			}
		case *ast.GenDecl:
			for _, spec := range typed.Specs {
				switch specTyped := spec.(type) {
				case *ast.TypeSpec:
					if specTyped.Name != nil && specTyped.Name.Name == name {
						return specTyped.Name
					}
				case *ast.ValueSpec:
					for _, ident := range specTyped.Names {
						if ident != nil && ident.Name == name {
							return ident
						}
					}
				}
			}
		}
	}
	return nil
}
