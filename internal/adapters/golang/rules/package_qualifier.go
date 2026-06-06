package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

	adapterproto "refactorlah/internal/adapters/contract"
)

const PackageQualifierRuleName = "go.PackageQualifierRule"

type PackageQualifierMapping struct {
	OldImport  string
	NewImport  string
	OldPackage string
	NewPackage string
}

type PackageQualifierInput struct {
	File     string
	Mappings []PackageQualifierMapping
}

type PackageQualifierRule struct{}

func (r PackageQualifierRule) Collect(source []byte, input PackageQualifierInput) ([]adapterproto.Replacement, error) {
	if len(input.Mappings) == 0 {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, input.File, source, 0)
	if err != nil {
		return nil, err
	}

	var replacements []adapterproto.Replacement
	for _, mapping := range input.Mappings {
		if mapping.OldPackage == "" || mapping.OldPackage == mapping.NewPackage {
			continue
		}
		if !importsPathWithoutAlias(file, mapping.OldImport) {
			continue
		}
		if declaresIdentifier(file, mapping.OldPackage) {
			continue
		}
		replacements = append(replacements, packageQualifierReplacements(fileSet, input.File, file, mapping)...)
	}

	return replacements, nil
}

func importsPathWithoutAlias(file *ast.File, importPath string) bool {
	for _, importSpec := range file.Imports {
		specPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil || specPath != importPath {
			continue
		}
		return importSpec.Name == nil
	}
	return false
}

func declaresIdentifier(file *ast.File, name string) bool {
	declared := false
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil || declared {
			return false
		}

		switch typed := node.(type) {
		case *ast.ValueSpec:
			declared = containsIdent(typed.Names, name)
		case *ast.TypeSpec:
			declared = typed.Name != nil && typed.Name.Name == name
		case *ast.AssignStmt:
			if typed.Tok == token.DEFINE {
				declared = expressionListDeclares(typed.Lhs, name)
			}
		case *ast.RangeStmt:
			if typed.Tok == token.DEFINE {
				declared = expressionDeclares(typed.Key, name) || expressionDeclares(typed.Value, name)
			}
		case *ast.Field:
			declared = containsIdent(typed.Names, name)
		}

		return !declared
	})
	return declared
}

func containsIdent(idents []*ast.Ident, name string) bool {
	for _, ident := range idents {
		if ident != nil && ident.Name == name {
			return true
		}
	}
	return false
}

func expressionListDeclares(expressions []ast.Expr, name string) bool {
	for _, expression := range expressions {
		if expressionDeclares(expression, name) {
			return true
		}
	}
	return false
}

func expressionDeclares(expression ast.Expr, name string) bool {
	ident, ok := expression.(*ast.Ident)
	return ok && ident.Name == name
}

func packageQualifierReplacements(fileSet *token.FileSet, filePath string, file *ast.File, mapping PackageQualifierMapping) []adapterproto.Replacement {
	var replacements []adapterproto.Replacement
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		identifier, ok := selector.X.(*ast.Ident)
		if !ok || identifier.Name != mapping.OldPackage {
			return true
		}

		tokenFile := fileSet.File(identifier.Pos())
		replacements = append(replacements, adapterproto.Replacement{
			File:        filePath,
			Start:       tokenFile.Offset(identifier.Pos()),
			End:         tokenFile.Offset(identifier.End()),
			Replacement: mapping.NewPackage,
			Reason:      "go-package-qualifier",
			Rule:        PackageQualifierRuleName,
			Adapter:     "go",
		})

		return true
	})

	return replacements
}
