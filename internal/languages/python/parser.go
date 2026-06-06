//go:build cgo

package python

import (
	"refactorlah/internal/languages/treesitter"

	treeSitterPython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

func Parse(source []byte) (*treesitter.Document, error) {
	return treesitter.Parse(source, treesitter.NewLanguage("python", treeSitterPython.Language()))
}
