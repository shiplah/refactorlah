package rules

import (
	"go/parser"
	"go/token"

	adapterproto "refactorlah/internal/adapters/contract"
)

const PackageDeclarationRuleName = "go.PackageDeclarationRule"

type PackageDeclarationInput struct {
	File       string
	OldPackage string
	NewPackage string
}

type PackageDeclarationRule struct{}

func (r PackageDeclarationRule) Collect(source []byte, input PackageDeclarationInput) ([]adapterproto.Replacement, error) {
	if input.OldPackage == "" || input.OldPackage == input.NewPackage {
		return nil, nil
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, input.File, source, parser.PackageClauseOnly)
	if err != nil {
		return nil, err
	}
	if file.Name == nil || file.Name.Name != input.OldPackage {
		return nil, nil
	}

	tokenFile := fileSet.File(file.Name.Pos())
	return []adapterproto.Replacement{{
		File:        input.File,
		Start:       tokenFile.Offset(file.Name.Pos()),
		End:         tokenFile.Offset(file.Name.End()),
		Replacement: input.NewPackage,
		Reason:      "go-package-declaration",
		Rule:        PackageDeclarationRuleName,
		Adapter:     "go",
	}}, nil
}
