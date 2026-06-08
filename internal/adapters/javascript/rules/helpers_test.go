package rules_test

import (
	"sort"
	"strings"

	"refactorlah/internal/replacements"
)

func applyRuleReplacements(content string, items []replacements.Replacement) string {
	sort.Slice(items, func(left int, right int) bool {
		return items[left].Start > items[right].Start
	})

	builder := strings.Builder{}
	builder.WriteString(content)
	for _, item := range items {
		updated := builder.String()
		builder.Reset()
		builder.WriteString(updated[:item.Start])
		builder.WriteString(item.Replacement)
		builder.WriteString(updated[item.End:])
	}
	return builder.String()
}
