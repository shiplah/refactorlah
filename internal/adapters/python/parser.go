//go:build cgo

package python

import (
	"github.com/shiplah/refactorlah/internal/parsing/treesitter"

	treeSitterPython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

func Parse(source []byte) (*treesitter.Document, error) {
	return treesitter.Parse(source, treesitter.NewLanguage("python", treeSitterPython.Language()))
}
