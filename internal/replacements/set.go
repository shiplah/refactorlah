package replacements

import adapterproto "refactorlah/internal/adapters/contract"

type replacementKey struct {
	file        string
	start       int
	end         int
	replacement string
}

func Deduplicate(replacements []adapterproto.Replacement) []adapterproto.Replacement {
	seen := map[replacementKey]bool{}
	deduplicated := make([]adapterproto.Replacement, 0, len(replacements))
	for _, replacement := range replacements {
		key := replacementKey{
			file:        replacement.File,
			start:       replacement.Start,
			end:         replacement.End,
			replacement: replacement.Replacement,
		}
		if seen[key] {
			continue
		}

		seen[key] = true
		deduplicated = append(deduplicated, replacement)
	}

	return deduplicated
}
