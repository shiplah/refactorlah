//go:build cgo

package rules

import (
	"refactorlah/internal/adapters/php/syntax"
	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const ClassDeclarationRuleName = "php.ClassDeclarationRule"

type ClassDeclarationInput struct {
	File         string
	OldShortName string
	NewShortName string
}

type ClassDeclarationRule struct{}

func (r ClassDeclarationRule) Collect(document *treesitter.Document, input ClassDeclarationInput) []replacements.Replacement {
	if input.OldShortName == "" || input.OldShortName == input.NewShortName {
		return nil
	}

	var result []replacements.Replacement
	for _, node := range document.NodesByKind("class_declaration", "interface_declaration", "trait_declaration", "enum_declaration") {
		if !isTopLevelDeclaration(node) {
			continue
		}

		match, ok := syntax.DeclarationNameOffset(node.Text)
		if !ok || match.Name != input.OldShortName {
			continue
		}

		result = append(result, replacements.Replacement{
			File:        input.File,
			Start:       node.StartByte + match.Start,
			End:         node.StartByte + match.End,
			Replacement: input.NewShortName,
			Reason:      "php-class-declaration",
			Rule:        ClassDeclarationRuleName,
			Adapter:     "php",
		})
	}

	return result
}

func isTopLevelDeclaration(node treesitter.Node) bool {
	switch node.ParentKind() {
	case "program", "namespace_definition":
		return true
	default:
		return false
	}
}
