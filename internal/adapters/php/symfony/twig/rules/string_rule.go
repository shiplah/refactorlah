package rules

import (
	"regexp"
	"strings"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/replacements"
)

type StringReplacementRule interface {
	Collect(file string, content string, mapping adapterproto.PathMapping) []replacements.Replacement
}

type PatternRule struct {
	RuleName string
	Reason   string
	Patterns func(quotedReference string) []string
}

func (r PatternRule) Collect(file string, content string, mapping adapterproto.PathMapping) []replacements.Replacement {
	var result []replacements.Replacement
	for _, quotedReference := range quotedOldReferences(mapping.OldReference) {
		for _, pattern := range r.Patterns(quotedReference) {
			expression := regexp.MustCompile(pattern)
			for _, match := range expression.FindAllStringSubmatchIndex(content, -1) {
				if len(match) < 4 || match[2] < 0 || match[3] < match[2] {
					continue
				}
				result = append(result, replacements.Replacement{
					File:        file,
					Start:       match[2],
					End:         match[3],
					Replacement: replacementForQuotedReference(quotedReference, mapping.NewReference),
					Reason:      r.Reason,
					Rule:        r.RuleName,
					Adapter:     "php",
				})
			}
		}
	}

	return result
}

func quotedOldReferences(oldReference string) []string {
	return []string{"'" + oldReference + "'", `"` + oldReference + `"`}
}

func replacementForQuotedReference(quotedReference string, newReference string) string {
	quote := "'"
	if strings.HasPrefix(quotedReference, `"`) {
		quote = `"`
	}

	return quote + newReference + quote
}

func quotedReferencePattern(quotedReference string) string {
	return regexp.QuoteMeta(quotedReference)
}
