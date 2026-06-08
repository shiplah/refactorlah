//go:build cgo

package javascript

import (
	"strconv"
	"strings"

	"refactorlah/internal/parsing/treesitter"
)

func objectPropertyValue(object treesitter.SyntaxNode, property string) (treesitter.SyntaxNode, bool) {
	for _, child := range object.NamedChildren() {
		if child.Kind() != "pair" {
			continue
		}
		key, value, ok := pairKeyValue(child)
		if ok && keyName(key) == property {
			return value, true
		}
	}
	return treesitter.SyntaxNode{}, false
}

func pairKeyValue(pair treesitter.SyntaxNode) (treesitter.SyntaxNode, treesitter.SyntaxNode, bool) {
	key, keyOK := pair.ChildByFieldName("key")
	value, valueOK := pair.ChildByFieldName("value")
	if keyOK && valueOK {
		return key, value, true
	}

	children := pair.NamedChildren()
	if len(children) < 2 {
		return treesitter.SyntaxNode{}, treesitter.SyntaxNode{}, false
	}
	return children[0], children[1], true
}

func keyName(node treesitter.SyntaxNode) string {
	switch node.Kind() {
	case "property_identifier", "identifier":
		return node.Text()
	case "string":
		value, ok := stringLiteralValue(node)
		if ok {
			return value
		}
	}
	return ""
}

func stringLiteralValue(node treesitter.SyntaxNode) (string, bool) {
	text := node.Text()
	if len(text) < 2 {
		return "", false
	}
	switch text[0] {
	case '"':
		value, err := strconv.Unquote(text)
		return value, err == nil
	case '\'':
		return unquoteSingleQuotedString(text)
	default:
		return "", false
	}
}

func unquoteSingleQuotedString(text string) (string, bool) {
	if len(text) < 2 || text[0] != '\'' || text[len(text)-1] != '\'' {
		return "", false
	}

	var builder strings.Builder
	escaped := false
	for index := 1; index < len(text)-1; index++ {
		current := text[index]
		if escaped {
			switch current {
			case '\'', '"', '\\', '/':
				builder.WriteByte(current)
			case 'n':
				builder.WriteByte('\n')
			case 'r':
				builder.WriteByte('\r')
			case 't':
				builder.WriteByte('\t')
			default:
				return "", false
			}
			escaped = false
			continue
		}
		if current == '\\' {
			escaped = true
			continue
		}
		builder.WriteByte(current)
	}
	if escaped {
		return "", false
	}
	return builder.String(), true
}

func walkSyntaxNode(node treesitter.SyntaxNode, visit func(treesitter.SyntaxNode)) {
	visit(node)
	for _, child := range node.NamedChildren() {
		walkSyntaxNode(child, visit)
	}
}
