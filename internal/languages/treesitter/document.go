//go:build cgo

package treesitter

import (
	"fmt"
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
	treeSitterPHP "github.com/tree-sitter/tree-sitter-php/bindings/go"
	treeSitterPython "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

type Language struct {
	name    string
	pointer unsafe.Pointer
}

func PHP() Language {
	return Language{name: "php", pointer: treeSitterPHP.LanguagePHP()}
}

func Python() Language {
	return Language{name: "python", pointer: treeSitterPython.Language()}
}

type Document struct {
	source []byte
	tree   *sitter.Tree
}

type Node struct {
	Kind      string
	StartByte int
	EndByte   int
	Text      string
}

func Parse(source []byte, language Language) (*Document, error) {
	parser := sitter.NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(sitter.NewLanguage(language.pointer)); err != nil {
		return nil, fmt.Errorf("set %s tree-sitter language: %w", language.name, err)
	}

	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil, fmt.Errorf("parse %s source: tree-sitter returned no tree", language.name)
	}

	document := &Document{
		source: source,
		tree:   tree,
	}

	if document.RootHasError() {
		document.Close()
		return nil, fmt.Errorf("parse %s source: syntax tree contains errors", language.name)
	}

	return document, nil
}

func (d *Document) Close() {
	if d != nil && d.tree != nil {
		d.tree.Close()
		d.tree = nil
	}
}

func (d *Document) RootHasError() bool {
	return d.tree.RootNode().HasError()
}

func (d *Document) NodesByKind(kinds ...string) []Node {
	wanted := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		wanted[kind] = true
	}

	var nodes []Node
	d.walk(d.tree.RootNode(), func(node *sitter.Node) {
		if wanted[node.Kind()] {
			start, end := node.ByteRange()
			nodes = append(nodes, Node{
				Kind:      node.Kind(),
				StartByte: int(start),
				EndByte:   int(end),
				Text:      node.Utf8Text(d.source),
			})
		}
	})

	return nodes
}

func (d *Document) walk(node *sitter.Node, visit func(*sitter.Node)) {
	visit(node)

	cursor := node.Walk()
	defer cursor.Close()

	for _, child := range node.NamedChildren(cursor) {
		childNode := child
		d.walk(&childNode, visit)
	}
}
