package names

import "strings"

func Short(symbol string) string {
	index := strings.LastIndex(symbol, "\\")
	if index < 0 {
		return symbol
	}
	return symbol[index+1:]
}

func Namespace(symbol string) string {
	index := strings.LastIndex(symbol, "\\")
	if index < 0 {
		return ""
	}
	return symbol[:index]
}

func ContainsIdentifier(content string, identifier string) bool {
	if identifier == "" {
		return false
	}

	offset := 0
	for {
		index := strings.Index(content[offset:], identifier)
		if index < 0 {
			return false
		}

		start := offset + index
		end := start + len(identifier)
		if IsNameBoundary(content, start-1) && IsNameBoundary(content, end) {
			return true
		}
		offset = end
	}
}

func IsNameBoundary(text string, index int) bool {
	if index < 0 || index >= len(text) {
		return true
	}

	return text[index] != '\\' && !IsIdentifierByte(text[index])
}

func IsIdentifierByte(value byte) bool {
	return value == '_' ||
		value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' ||
		value >= 0x80
}

func IsWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}
