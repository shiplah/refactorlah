package php

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
		for start < len(text) && isSpace(text[start]) {
			start++
		}

		end := start
		for end < len(text) && isIdentifierByte(text[end]) {
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
