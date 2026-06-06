//go:build cgo

package php

import (
	"os"
	"path/filepath"
)

func (s *SymbolScanner) primarySymbolKind(projectRoot string, relativePath string, expectedShortName string) (string, bool) {
	source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(relativePath)))
	if err != nil {
		return "", false
	}

	document, err := Parse(source)
	if err != nil {
		return "", false
	}
	defer document.Close()

	candidates := document.NodesByKind("class_declaration", "interface_declaration", "trait_declaration", "enum_declaration")
	var matchingKind string
	for _, candidate := range candidates {
		name := declarationName(candidate.Text)
		if name == expectedShortName {
			return phpSymbolKind(candidate.Kind), true
		}
		if matchingKind == "" {
			matchingKind = phpSymbolKind(candidate.Kind)
		}
	}

	if len(candidates) == 1 {
		return matchingKind, true
	}

	return "", false
}

func phpSymbolKind(nodeKind string) string {
	switch nodeKind {
	case "interface_declaration":
		return "interface"
	case "trait_declaration":
		return "trait"
	case "enum_declaration":
		return "enum"
	default:
		return "class"
	}
}

func declarationName(text string) string {
	keywords := []string{"class", "interface", "trait", "enum"}
	for _, keyword := range keywords {
		name, ok := wordAfterKeyword(text, keyword)
		if ok {
			return name
		}
	}
	return ""
}

func wordAfterKeyword(text string, keyword string) (string, bool) {
	index := 0
	for {
		found := indexOfWord(text[index:], keyword)
		if found < 0 {
			return "", false
		}
		start := index + found + len(keyword)
		for start < len(text) && isSpace(text[start]) {
			start++
		}
		end := start
		for end < len(text) && isIdentifierByte(text[end]) {
			end++
		}
		if end > start {
			return text[start:end], true
		}
		index = start
	}
}

func indexOfWord(text string, word string) int {
	for index := 0; index+len(word) <= len(text); index++ {
		if text[index:index+len(word)] != word {
			continue
		}
		before := index - 1
		after := index + len(word)
		if (before < 0 || !isIdentifierByte(text[before])) && (after >= len(text) || !isIdentifierByte(text[after])) {
			return index
		}
	}
	return -1
}

func isSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func isIdentifierByte(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
