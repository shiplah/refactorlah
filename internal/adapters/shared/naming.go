package shared

import "unicode"

func UpperFirst(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func LowerFirst(value string) string {
	if value == "" {
		return ""
	}

	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
