//go:build cgo

package rules

import (
	"strings"

	"github.com/NickSdot/refactorlah/internal/parsing/treesitter"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

const NamespaceDeclarationRuleName = "php.NamespaceDeclarationRule"

type NamespaceDeclarationInput struct {
	File         string
	OldNamespace string
	NewNamespace string
}

type NamespaceDeclarationRule struct{}

func (r NamespaceDeclarationRule) Collect(document *treesitter.Document, input NamespaceDeclarationInput) []replacements.Replacement {
	if input.OldNamespace == "" || input.OldNamespace == input.NewNamespace {
		return nil
	}

	for _, node := range document.NodesByKind("namespace_definition") {
		namespaceStart := strings.Index(node.Text, input.OldNamespace)
		if namespaceStart < 0 {
			continue
		}

		start := node.StartByte + namespaceStart
		return []replacements.Replacement{{
			File:        input.File,
			Start:       start,
			End:         start + len(input.OldNamespace),
			Replacement: input.NewNamespace,
			Reason:      "php-namespace-declaration",
			Rule:        NamespaceDeclarationRuleName,
			Adapter:     "php",
		}}
	}

	return nil
}
