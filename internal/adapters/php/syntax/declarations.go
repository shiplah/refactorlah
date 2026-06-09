package syntax

import "github.com/shiplah/refactorlah/internal/adapters/php/names"

type DeclarationNameMatch struct {
	Name  string
	Start int
	End   int
}

func DeclarationName(text string) string {
	match, ok := DeclarationNameOffset(text)
	if !ok {
		return ""
	}
	return match.Name
}

func DeclarationNameOffset(text string) (DeclarationNameMatch, bool) {
	keywords := []string{"class", "interface", "trait", "enum"}
	for _, keyword := range keywords {
		match, ok := wordAfterKeyword(text, keyword)
		if ok {
			return match, true
		}
	}
	return DeclarationNameMatch{}, false
}

func wordAfterKeyword(text string, keyword string) (DeclarationNameMatch, bool) {
	index := 0
	for {
		found := indexOfWord(text[index:], keyword)
		if found < 0 {
			return DeclarationNameMatch{}, false
		}

		start := index + found + len(keyword)
		for start < len(text) && names.IsWhitespace(text[start]) {
			start++
		}

		end := start
		for end < len(text) && names.IsIdentifierByte(text[end]) {
			end++
		}
		if end > start {
			return DeclarationNameMatch{
				Name:  text[start:end],
				Start: start,
				End:   end,
			}, true
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
		if (before < 0 || !names.IsIdentifierByte(text[before])) && (after >= len(text) || !names.IsIdentifierByte(text[after])) {
			return index
		}
	}

	return -1
}
