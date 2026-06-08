//go:build cgo

package javascript

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

const (
	viteAliasReason    = "javascript-vite-alias"
	viteAliasRule      = "javascript.ViteAliasRule"
	webpackAliasReason = "javascript-webpack-alias"
	webpackAliasRule   = "javascript.WebpackAliasRule"
)

type bundlerAliasConfig struct {
	fileNames []string
	reason    string
	rule      string
	label     string
}

var bundlerAliasConfigs = []bundlerAliasConfig{
	{
		fileNames: []string{"vite.config.js", "vite.config.ts", "vite.config.mjs", "vite.config.cjs"},
		reason:    viteAliasReason,
		rule:      viteAliasRule,
		label:     "Vite",
	},
	{
		fileNames: []string{"webpack.config.js", "webpack.config.ts", "webpack.config.mjs", "webpack.config.cjs"},
		reason:    webpackAliasReason,
		rule:      webpackAliasRule,
		label:     "webpack",
	},
}

func (a *Analyzer) collectBundlerAliasReplacements(projectRoot string, plan planning.MovePlan, scanIndex *scan.Index) ([]replacements.Replacement, []adapterproto.Warning, error) {
	var allReplacements []replacements.Replacement
	var warnings []adapterproto.Warning
	for _, config := range bundlerAliasConfigs {
		for _, file := range config.fileNames {
			mappings, configWarnings, err := readBundlerAliasMappings(projectRoot, file, config)
			if err != nil {
				return nil, nil, err
			}
			warnings = append(warnings, configWarnings...)
			rewrites := specifierRewritesForPathAliases(mappings, plan.Moves, config.reason, config.rule)
			if len(rewrites) == 0 {
				continue
			}

			files, err := scanIndex.CandidateFiles(projectRoot, specifierRewriteCandidateQuery(rewrites))
			if err != nil {
				return nil, nil, err
			}
			configReplacements, err := a.scanner.ScanSpecifiers(projectRoot, files, rewrites)
			if err != nil {
				return nil, nil, err
			}
			allReplacements = append(allReplacements, configReplacements...)
		}
	}
	return allReplacements, warnings, nil
}

func readBundlerAliasMappings(projectRoot string, file string, config bundlerAliasConfig) ([]pathAliasMapping, []adapterproto.Warning, error) {
	configPath := filepath.Join(projectRoot, file)
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	document, err := parseJavaScriptConfig(file, content)
	if err != nil {
		return nil, []adapterproto.Warning{{
			File:    file,
			Message: config.label + " config could not be parsed; alias rewrites skipped.",
		}}, nil
	}
	defer document.Close()

	mappings := bundlerAliasMappingsFromDocument(projectRoot, document.RootNode(), config)
	return mappings, nil, nil
}

func bundlerAliasMappingsFromDocument(projectRoot string, root treesitter.SyntaxNode, config bundlerAliasConfig) []pathAliasMapping {
	seen := map[pathAliasMapping]bool{}
	var mappings []pathAliasMapping
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
		if mappings[left].aliasPrefix == mappings[right].aliasPrefix {
			return mappings[left].targetPrefix < mappings[right].targetPrefix
		}
		return mappings[left].aliasPrefix < mappings[right].aliasPrefix
	})
	return mappings
}

func aliasMappings(projectRoot string, aliasNode treesitter.SyntaxNode, config bundlerAliasConfig) []pathAliasMapping {
	switch aliasNode.Kind() {
	case "object":
		return aliasObjectMappings(projectRoot, aliasNode)
	case "array":
		if config.reason != viteAliasReason {
			return nil
		}
		return viteAliasArrayMappings(projectRoot, aliasNode)
	default:
		return nil
	}
}

func aliasObjectMappings(projectRoot string, aliasObject treesitter.SyntaxNode) []pathAliasMapping {
	var mappings []pathAliasMapping
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

func viteAliasArrayMappings(projectRoot string, aliasArray treesitter.SyntaxNode) []pathAliasMapping {
	var mappings []pathAliasMapping
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

func bundlerAliasMapping(projectRoot string, alias string, replacement string) (pathAliasMapping, bool) {
	if alias == "" || strings.Contains(alias, "*") || !filepath.IsAbs(replacement) {
		return pathAliasMapping{}, false
	}

	relative, err := filepath.Rel(projectRoot, replacement)
	if err != nil {
		return pathAliasMapping{}, false
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || filepath.IsAbs(relative) || startsWithParentTraversal(relative) {
		return pathAliasMapping{}, false
	}

	aliasPrefix := alias
	if !strings.HasSuffix(aliasPrefix, "/") {
		aliasPrefix += "/"
	}
	targetPrefix := ""
	if relative != "." {
		targetPrefix = strings.TrimSuffix(relative, "/") + "/"
	}
	return pathAliasMapping{
		aliasPrefix:  aliasPrefix,
		targetPrefix: targetPrefix,
	}, true
}

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
