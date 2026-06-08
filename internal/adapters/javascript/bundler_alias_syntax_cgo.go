//go:build cgo

package javascript

import (
	"sort"

	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/parsing/treesitter"
)

func bundlerAliasMappingsFromDocument(projectRoot string, root treesitter.SyntaxNode, config bundlerAliasConfig) []rules.PathAliasMapping {
	seen := map[rules.PathAliasMapping]bool{}
	var mappings []rules.PathAliasMapping
	walkSyntaxNode(root, func(node treesitter.SyntaxNode) {
		if node.Kind() != "pair" {
			return
		}
		key, value, ok := pairKeyValue(node)
		if !ok || keyName(key) != "resolve" || value.Kind() != "object" {
			return
		}

		aliasNode, ok := objectPropertyValue(value, "alias")
		if !ok {
			return
		}
		for _, mapping := range aliasMappings(projectRoot, aliasNode, config) {
			if seen[mapping] {
				continue
			}
			seen[mapping] = true
			mappings = append(mappings, mapping)
		}
	})
	sort.Slice(mappings, func(left int, right int) bool {
		if mappings[left].AliasPrefix == mappings[right].AliasPrefix {
			return mappings[left].TargetPrefix < mappings[right].TargetPrefix
		}
		return mappings[left].AliasPrefix < mappings[right].AliasPrefix
	})
	return mappings
}

func aliasMappings(projectRoot string, aliasNode treesitter.SyntaxNode, config bundlerAliasConfig) []rules.PathAliasMapping {
	switch aliasNode.Kind() {
	case "object":
		return aliasObjectMappings(projectRoot, aliasNode)
	case "array":
		if config.reason != rules.ViteAliasReason {
			return nil
		}
		return viteAliasArrayMappings(projectRoot, aliasNode)
	default:
		return nil
	}
}

func aliasObjectMappings(projectRoot string, aliasObject treesitter.SyntaxNode) []rules.PathAliasMapping {
	var mappings []rules.PathAliasMapping
	for _, child := range aliasObject.NamedChildren() {
		if child.Kind() != "pair" {
			continue
		}
		key, value, ok := pairKeyValue(child)
		if !ok || value.Kind() != "string" {
			continue
		}
		alias := keyName(key)
		replacement, ok := stringLiteralValue(value)
		if !ok {
			continue
		}
		if mapping, ok := bundlerAliasMapping(projectRoot, alias, replacement); ok {
			mappings = append(mappings, mapping)
		}
	}
	return mappings
}

func viteAliasArrayMappings(projectRoot string, aliasArray treesitter.SyntaxNode) []rules.PathAliasMapping {
	var mappings []rules.PathAliasMapping
	for _, child := range aliasArray.NamedChildren() {
		if child.Kind() != "object" {
			continue
		}
		findNode, findOK := objectPropertyValue(child, "find")
		replacementNode, replacementOK := objectPropertyValue(child, "replacement")
		if !findOK || !replacementOK || findNode.Kind() != "string" || replacementNode.Kind() != "string" {
			continue
		}
		alias, aliasOK := stringLiteralValue(findNode)
		replacement, replacementOK := stringLiteralValue(replacementNode)
		if !aliasOK || !replacementOK {
			continue
		}
		if mapping, ok := bundlerAliasMapping(projectRoot, alias, replacement); ok {
			mappings = append(mappings, mapping)
		}
	}
	return mappings
}
