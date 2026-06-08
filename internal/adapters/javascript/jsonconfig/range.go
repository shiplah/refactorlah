package jsonconfig

import "strconv"

type Range struct {
	start int
	end   int
}

func ObjectPropertyRange(content []byte, propertyName string) (Range, bool) {
	return objectPropertyRangeBetween(content, Range{start: 0, end: len(content)}, propertyName)
}

func ObjectPropertyRangeIn(content []byte, objectRange Range, propertyName string) (Range, bool) {
	return objectPropertyRangeBetween(content, objectRange, propertyName)
}

func objectPropertyRangeBetween(content []byte, searchRange Range, propertyName string) (Range, bool) {
	depth := 0
	for index := searchRange.start; index < searchRange.end; {
		switch content[index] {
		case '"':
			key, _, _, next, ok := stringToken(content, index)
			if !ok {
				return Range{}, false
			}
			if depth == 1 && key == propertyName {
				colon := nextNonWhitespace(content, next)
				if colon < len(content) && content[colon] == ':' {
					objectStart := nextNonWhitespace(content, colon+1)
					if objectStart < len(content) && content[objectStart] == '{' {
						objectEnd, ok := matchingObjectEnd(content, objectStart)
						return Range{start: objectStart, end: objectEnd}, ok
					}
				}
			}
			index = next
		case '{':
			depth++
			index++
		case '}':
			depth--
			index++
		case '/':
			if next, ok := commentEnd(content, index); ok {
				index = next
			} else {
				index++
			}
		default:
			index++
		}
	}
	return Range{}, false
}

func matchingObjectEnd(content []byte, objectStart int) (int, bool) {
	depth := 0
	for index := objectStart; index < len(content); {
		switch content[index] {
		case '"':
			_, _, _, next, ok := stringToken(content, index)
			if !ok {
				return 0, false
			}
			index = next
		case '{':
			depth++
			index++
		case '}':
			depth--
			index++
			if depth == 0 {
				return index, true
			}
		case '/':
			if next, ok := commentEnd(content, index); ok {
				index = next
			} else {
				index++
			}
		default:
			index++
		}
	}
	return 0, false
}

func stringToken(content []byte, quoteStart int) (string, int, int, int, bool) {
	if quoteStart >= len(content) || content[quoteStart] != '"' {
		return "", 0, 0, 0, false
	}

	escaped := false
	for index := quoteStart + 1; index < len(content); index++ {
		if escaped {
			escaped = false
			continue
		}
		if content[index] == '\\' {
			escaped = true
			continue
		}
		if content[index] == '"' {
			decoded, err := strconv.Unquote(string(content[quoteStart : index+1]))
			if err != nil {
				return "", 0, 0, 0, false
			}
			return decoded, quoteStart + 1, index, index + 1, true
		}
	}
	return "", 0, 0, 0, false
}

func nextNonWhitespace(content []byte, index int) int {
	for index < len(content) {
		switch content[index] {
		case ' ', '\t', '\n', '\r':
			index++
		case '/':
			if next, ok := commentEnd(content, index); ok {
				index = next
			} else {
				return index
			}
		default:
			return index
		}
	}
	return len(content)
}

func commentEnd(content []byte, index int) (int, bool) {
	if index+1 >= len(content) || content[index] != '/' {
		return 0, false
	}
	switch content[index+1] {
	case '/':
		index += 2
		for index < len(content) && content[index] != '\n' {
			index++
		}
		return index, true
	case '*':
		index += 2
		for index+1 < len(content) {
			if content[index] == '*' && content[index+1] == '/' {
				return index + 2, true
			}
			index++
		}
		return len(content), true
	default:
		return 0, false
	}
}
