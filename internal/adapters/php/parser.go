//go:build cgo

package php

import (
	"refactorlah/internal/parsing/treesitter"

	treeSitterPHP "github.com/tree-sitter/tree-sitter-php/bindings/go"
)

func Parse(source []byte) (*treesitter.Document, error) {
	return treesitter.Parse(source, treesitter.NewLanguage("php", treeSitterPHP.LanguagePHP()))
}
