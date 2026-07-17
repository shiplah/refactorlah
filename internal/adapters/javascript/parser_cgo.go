//go:build cgo

package javascript

import (
	"path/filepath"

	"github.com/shiplah/refactorlah/internal/parsing/treesitter"

	treeSitterJavaScript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	treeSitterTypeScript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func parseJavaScriptConfig(path string, source []byte) (*treesitter.Document, error) {
	switch filepath.Ext(path) {
	case ".ts":
		return treesitter.Parse(source, treesitter.NewLanguage("typescript", treeSitterTypeScript.LanguageTypescript()))
	default:
		return treesitter.Parse(source, treesitter.NewLanguage("javascript", treeSitterJavaScript.Language()))
	}
}
