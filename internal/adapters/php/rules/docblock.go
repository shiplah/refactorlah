//go:build cgo

package rules

import (
	"sort"
	"strconv"
	"strings"

	"refactorlah/internal/adapters/php/names"
	"refactorlah/internal/parsing/treesitter"
	"refactorlah/internal/replacements"
)

const (
	DocblockVarRuleName    = "php.DocblockVarRule"
	DocblockParamRuleName  = "php.DocblockParamRule"
	DocblockReturnRuleName = "php.DocblockReturnRule"
	DocblockThrowsRuleName = "php.DocblockThrowsRule"
)

type DocblockVarRule struct{}

func (r DocblockVarRule) Collect(document *treesitter.Document, input SymbolReferenceInput) []replacements.Replacement {
	return collectDocblockTagReplacements(document, input, "var", "php-docblock-var", DocblockVarRuleName)
}

type DocblockParamRule struct{}

func (r DocblockParamRule) Collect(document *treesitter.Document, input SymbolReferenceInput) []replacements.Replacement {
	return collectDocblockTagReplacements(document, input, "param", "php-docblock-param", DocblockParamRuleName)
}

type DocblockReturnRule struct{}

func (r DocblockReturnRule) Collect(document *treesitter.Document, input SymbolReferenceInput) []replacements.Replacement {
	return collectDocblockTagReplacements(document, input, "return", "php-docblock-return", DocblockReturnRuleName)
}

type DocblockThrowsRule struct{}

func (r DocblockThrowsRule) Collect(document *treesitter.Document, input SymbolReferenceInput) []replacements.Replacement {
	return collectDocblockTagReplacements(document, input, "throws", "php-docblock-throws", DocblockThrowsRuleName)
}

type docblockReferenceReplacement struct {
	old string
	new string
}

type docblockLineRange struct {
	start int
	end   int
}

func collectDocblockTagReplacements(document *treesitter.Document, input SymbolReferenceInput, tag string, reason string, rule string) []replacements.Replacement {
	if input.OldSymbol == "" || input.OldSymbol == input.NewSymbol {
		return nil
	}

	referenceReplacements := docblockReferenceReplacements(document, input)
	if len(referenceReplacements) == 0 {
		return nil
	}

	var result []replacements.Replacement
	seen := map[string]bool{}
	for _, comment := range document.NodesByKind("comment") {
		for _, lineRange := range docblockLineRanges(comment.Text) {
			line := comment.Text[lineRange.start:lineRange.end]
			if !containsDocblockTag(line, tag) {
				continue
			}

			for _, candidate := range referenceReplacements {
				for _, matchStart := range findDocblockReferenceMatches(line, candidate.old) {
					start := comment.StartByte + lineRange.start + matchStart
					end := start + len(candidate.old)
					key := replacementRangeKey(start, end)
					if seen[key] {
						continue
					}

					result = append(result, replacements.Replacement{
						File:        input.File,
						Start:       start,
						End:         end,
						Replacement: candidate.new,
						Reason:      reason,
						Rule:        rule,
						Adapter:     "php",
					})
					seen[key] = true
				}
			}
		}
	}

	return result
}

func docblockReferenceReplacements(document *treesitter.Document, input SymbolReferenceInput) []docblockReferenceReplacement {
	replacementsByOld := map[string]string{
		strings.TrimPrefix(input.OldSymbol, "\\"):        strings.TrimPrefix(input.NewSymbol, "\\"),
		"\\" + strings.TrimPrefix(input.OldSymbol, "\\"): "\\" + strings.TrimPrefix(input.NewSymbol, "\\"),
	}

	oldShort := names.Short(input.OldSymbol)
	newShort := names.Short(input.NewSymbol)
	if importedReference, ok := importedShortReplacement(document, input.OldSymbol, input.NewSymbol, oldShort); ok && importedReference != oldShort {
		replacementsByOld[oldShort] = importedReference
	}

	if input.OldNamespace != "" && declaredNamespace(document) == input.OldNamespace && oldShort != newShort {
		replacementsByOld[oldShort] = newShort
	}

	result := make([]docblockReferenceReplacement, 0, len(replacementsByOld))
	for oldReference, newReference := range replacementsByOld {
		if oldReference == "" || oldReference == newReference {
			continue
		}
		result = append(result, docblockReferenceReplacement{old: oldReference, new: newReference})
	}
	sort.Slice(result, func(left int, right int) bool {
		return len(result[left].old) > len(result[right].old)
	})

	return result
}

func docblockLineRanges(text string) []docblockLineRange {
	var ranges []docblockLineRange
	start := 0
	for start < len(text) {
		end := start
		for end < len(text) && text[end] != '\n' && text[end] != '\r' {
			end++
		}
		ranges = append(ranges, docblockLineRange{start: start, end: end})

		start = end
		for start < len(text) && (text[start] == '\n' || text[start] == '\r') {
			start++
		}
	}

	return ranges
}

func containsDocblockTag(line string, tag string) bool {
	token := "@" + tag
	offset := 0
	for {
		index := strings.Index(line[offset:], token)
		if index < 0 {
			return false
		}

		start := offset + index
		end := start + len(token)
		if isDocblockTagBoundary(line, end) {
			return true
		}
		offset = end
	}
}

func findDocblockReferenceMatches(line string, oldReference string) []int {
	var starts []int
	offset := 0
	for {
		index := strings.Index(line[offset:], oldReference)
		if index < 0 {
			return starts
		}

		start := offset + index
		end := start + len(oldReference)
		if isDocblockSymbolBoundary(line, start-1) && isDocblockSymbolBoundary(line, end) {
			starts = append(starts, start)
		}
		offset = end
	}
}

func isDocblockTagBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	return !names.IsIdentifierByte(text[index])
}

func isDocblockSymbolBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	return text[index] != '\\' && !names.IsIdentifierByte(text[index])
}

func replacementRangeKey(start int, end int) string {
	return strconv.Itoa(start) + ":" + strconv.Itoa(end)
}
