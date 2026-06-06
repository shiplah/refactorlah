//go:build cgo

package php

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/php/names"
	"refactorlah/internal/adapters/shared"
)

type SemanticHintScanner struct{}

func (s SemanticHintScanner) Scan(projectRoot string, phpFiles []string, textFiles []string, mappings []adapterproto.SymbolMapping) ([]adapterproto.Warning, error) {
	var warnings []adapterproto.Warning
	for _, file := range phpFiles {
		content, err := readSemanticHintFile(projectRoot, file)
		if err != nil {
			return nil, err
		}
		warnings = append(warnings, s.scanPHPFile(file, content, mappings)...)
	}

	for _, file := range textFiles {
		content, err := readSemanticHintFile(projectRoot, file)
		if err != nil {
			return nil, err
		}
		warnings = append(warnings, s.scanTextFile(file, content, mappings)...)
	}

	return deduplicateSemanticWarnings(warnings), nil
}

func (s SemanticHintScanner) scanPHPFile(file string, content string, mappings []adapterproto.SymbolMapping) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, mapping := range mappings {
		for oldName, newName := range variableHints(mapping) {
			pattern := regexp.MustCompile(`\$` + regexp.QuoteMeta(oldName) + `\b`)
			for _, match := range pattern.FindAllStringIndex(content, -1) {
				warnings = append(warnings, semanticWarning(file, content, match[0], oldName, newName))
			}
		}

		for oldLiteral, newLiteral := range literalHints(mapping) {
			warnings = append(warnings, phpStringLiteralWarnings(file, content, oldLiteral, newLiteral)...)
			warnings = append(warnings, suspiciousNameWarnings(file, content, mapping, oldLiteral, newLiteral)...)
		}
	}
	return warnings
}

func (s SemanticHintScanner) scanTextFile(file string, content string, mappings []adapterproto.SymbolMapping) []adapterproto.Warning {
	var warnings []adapterproto.Warning
	for _, mapping := range mappings {
		for oldLiteral, newLiteral := range literalHints(mapping) {
			offset := strings.Index(content, oldLiteral)
			if offset < 0 {
				continue
			}
			warnings = append(warnings, semanticWarning(file, content, offset, oldLiteral, newLiteral))
		}
	}
	return warnings
}

func phpStringLiteralWarnings(file string, content string, oldLiteral string, newLiteral string) []adapterproto.Warning {
	pattern := regexp.MustCompile(`(['"])([^'"]*` + regexp.QuoteMeta(oldLiteral) + `[^'"]*)(['"])`)
	var warnings []adapterproto.Warning
	for _, match := range pattern.FindAllStringSubmatchIndex(content, -1) {
		if len(match) < 8 || match[4] < 0 || match[5] < match[4] {
			continue
		}
		if content[match[2]:match[3]] != content[match[6]:match[7]] {
			continue
		}

		value := content[match[4]:match[5]]
		warnings = append(warnings, semanticWarning(file, content, match[4], oldLiteral, strings.ReplaceAll(value, oldLiteral, newLiteral)))
	}
	return warnings
}

func suspiciousNameWarnings(file string, content string, mapping adapterproto.SymbolMapping, oldLiteral string, newLiteral string) []adapterproto.Warning {
	oldShortName := names.Short(mapping.OldSymbol)
	if oldShortName == "" || oldLiteral != shared.LowerFirst(oldShortName) {
		return nil
	}

	pattern := regexp.MustCompile(`\b[A-Z][A-Za-z0-9_]*` + regexp.QuoteMeta(oldShortName) + `[A-Za-z0-9_]*\b`)
	var warnings []adapterproto.Warning
	for _, match := range pattern.FindAllStringIndex(content, -1) {
		reference := content[match[0]:match[1]]
		if reference == oldShortName {
			continue
		}
		warnings = append(warnings, semanticWarning(file, content, match[0], reference, strings.ReplaceAll(reference, oldShortName, shared.UpperFirst(newLiteral))))
	}
	return warnings
}

func semanticWarning(file string, content string, offset int, oldName string, newName string) adapterproto.Warning {
	return adapterproto.Warning{
		File:    file,
		Line:    strings.Count(content[:offset], "\n") + 1,
		Message: `Semantic name "` + oldName + `" resembles moved symbol; consider "` + newName + `". Not changed.`,
	}
}

func readSemanticHintFile(projectRoot string, file string) (string, error) {
	content, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func variableHints(mapping adapterproto.SymbolMapping) map[string]string {
	oldLowerCamel := shared.LowerFirst(names.Short(mapping.OldSymbol))
	newLowerCamel := shared.LowerFirst(names.Short(mapping.NewSymbol))
	return map[string]string{
		oldLowerCamel:       newLowerCamel,
		oldLowerCamel + "s": newLowerCamel + "s",
	}
}

func literalHints(mapping adapterproto.SymbolMapping) map[string]string {
	oldShortName := names.Short(mapping.OldSymbol)
	newShortName := names.Short(mapping.NewSymbol)
	oldSnake := toDelimited(oldShortName, "_")
	newSnake := toDelimited(newShortName, "_")
	oldKebab := toDelimited(oldShortName, "-")
	newKebab := toDelimited(newShortName, "-")

	return map[string]string{
		shared.LowerFirst(oldShortName):       shared.LowerFirst(newShortName),
		shared.LowerFirst(oldShortName) + "s": shared.LowerFirst(newShortName) + "s",
		oldSnake:                              newSnake,
		oldSnake + "s":                        newSnake + "s",
		oldKebab:                              newKebab,
		oldKebab + "s":                        newKebab + "s",
	}
}

func toDelimited(name string, delimiter string) string {
	if name == "" {
		return ""
	}

	var words []string
	start := 0
	for index := 1; index < len(name); index++ {
		current := rune(name[index])
		previous := rune(name[index-1])
		nextLower := index+1 < len(name) && unicode.IsLower(rune(name[index+1]))
		if unicode.IsUpper(current) && (unicode.IsLower(previous) || nextLower) {
			words = append(words, strings.ToLower(name[start:index]))
			start = index
		}
	}
	words = append(words, strings.ToLower(name[start:]))
	return strings.Join(words, delimiter)
}

func deduplicateSemanticWarnings(warnings []adapterproto.Warning) []adapterproto.Warning {
	seen := map[string]bool{}
	var result []adapterproto.Warning
	for _, warning := range warnings {
		key := warning.File + ":" + warning.Message + ":" + strconv.Itoa(warning.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, warning)
	}
	return result
}
