//go:build cgo

package treesitter

import (
	"fmt"
	"unsafe"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type Language struct {
	name    string
	pointer unsafe.Pointer
}

func NewLanguage(name string, pointer unsafe.Pointer) Language {
	return Language{name: name, pointer: pointer}
}

type Document struct {
	source []byte
	tree   *sitter.Tree
}

type Node struct {
	Kind          string
	StartByte     int
	EndByte       int
	Text          string
	AncestorKinds []string
}

type SyntaxNode struct {
	source []byte
	node   *sitter.Node
}

func Parse(source []byte, language Language) (*Document, error) {
	return parse(source, language, false)
}

func ParseRecovering(source []byte, language Language) (*Document, error) {
	return parse(source, language, true)
}

func parse(source []byte, language Language, allowErrors bool) (*Document, error) {
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
		if allowErrors {
			return document, nil
		}
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

func (d *Document) RootNode() SyntaxNode {
	return SyntaxNode{
		source: d.source,
		node:   d.tree.RootNode(),
	}
}

func (n SyntaxNode) Kind() string {
	if n.node == nil {
		return ""
	}
	return n.node.Kind()
}

func (n SyntaxNode) Text() string {
	if n.node == nil {
		return ""
	}
	return n.node.Utf8Text(n.source)
}

func (n SyntaxNode) StartByte() int {
	if n.node == nil {
		return 0
	}
	return int(n.node.StartByte())
}

func (n SyntaxNode) EndByte() int {
	if n.node == nil {
		return 0
	}
	return int(n.node.EndByte())
}

func (n SyntaxNode) ChildByFieldName(fieldName string) (SyntaxNode, bool) {
	if n.node == nil {
		return SyntaxNode{}, false
	}
	child := n.node.ChildByFieldName(fieldName)
	if child == nil {
		return SyntaxNode{}, false
	}
	return SyntaxNode{source: n.source, node: child}, true
}

func (n SyntaxNode) NamedChildren() []SyntaxNode {
	if n.node == nil {
		return nil
	}

	cursor := n.node.Walk()
	defer cursor.Close()

	children := n.node.NamedChildren(cursor)
	result := make([]SyntaxNode, 0, len(children))
	for index := range children {
		child := children[index]
		result = append(result, SyntaxNode{source: n.source, node: &child})
	}
	return result
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
				Kind:          node.Kind(),
				StartByte:     int(start),
				EndByte:       int(end),
				Text:          node.Utf8Text(d.source),
				AncestorKinds: ancestorKinds(node),
			})
		}
	})

	return nodes
}

func (n Node) ParentKind() string {
	if len(n.AncestorKinds) == 0 {
		return ""
	}

	return n.AncestorKinds[0]
}

func ancestorKinds(node *sitter.Node) []string {
	var kinds []string
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		kinds = append(kinds, parent.Kind())
	}

	return kinds
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
