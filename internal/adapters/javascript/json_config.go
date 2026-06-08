package javascript

import (
	"strconv"

	"refactorlah/internal/replacements"
)

type jsonRange struct {
	start int
	end   int
}

func jsonObjectPropertyRange(content []byte, propertyName string) (jsonRange, bool) {
	return jsonObjectPropertyRangeBetween(content, jsonRange{start: 0, end: len(content)}, propertyName)
}

func jsonObjectPropertyRangeIn(content []byte, objectRange jsonRange, propertyName string) (jsonRange, bool) {
	return jsonObjectPropertyRangeBetween(content, objectRange, propertyName)
}

func jsonObjectPropertyRangeBetween(content []byte, searchRange jsonRange, propertyName string) (jsonRange, bool) {
	depth := 0
	for index := searchRange.start; index < searchRange.end; {
		switch content[index] {
		case '"':
			key, _, _, next, ok := jsonStringToken(content, index)
			if !ok {
				return jsonRange{}, false
			}
			if depth == 1 && key == propertyName {
				colon := nextNonJSONWhitespace(content, next)
				if colon < len(content) && content[colon] == ':' {
					objectStart := nextNonJSONWhitespace(content, colon+1)
					if objectStart < len(content) && content[objectStart] == '{' {
						objectEnd, ok := matchingJSONObjectEnd(content, objectStart)
						return jsonRange{start: objectStart, end: objectEnd}, ok
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
			if next, ok := jsonCommentEnd(content, index); ok {
				index = next
			} else {
				index++
			}
		default:
			index++
		}
	}
	return jsonRange{}, false
}

func matchingJSONObjectEnd(content []byte, objectStart int) (int, bool) {
	depth := 0
	for index := objectStart; index < len(content); {
		switch content[index] {
		case '"':
			_, _, _, next, ok := jsonStringToken(content, index)
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
			if next, ok := jsonCommentEnd(content, index); ok {
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

func jsonObjectStringValueReplacements(file string, content []byte, objectRange jsonRange, rewrites map[string]string, reason string, rule string) []replacements.Replacement {
	var result []replacements.Replacement
	depth := 1
	for index := objectRange.start + 1; index < objectRange.end-1; {
		switch content[index] {
		case '"':
			_, _, _, next, ok := jsonStringToken(content, index)
			if !ok {
				return nil
			}
			if depth == 1 {
				colon := nextNonJSONWhitespace(content, next)
				valueStart := nextNonJSONWhitespace(content, colon+1)
				if colon < objectRange.end && content[colon] == ':' && valueStart < objectRange.end && content[valueStart] == '"' {
					value, rawStart, rawEnd, valueNext, ok := jsonStringToken(content, valueStart)
					if !ok {
						return nil
					}
					rawValue := string(content[rawStart:rawEnd])
					if value == rawValue {
						if replacement, ok := rewrites[value]; ok {
							result = append(result, replacements.Replacement{
								File:        file,
								Start:       rawStart,
								End:         rawEnd,
								Replacement: replacement,
								Reason:      reason,
								Rule:        rule,
								Adapter:     "javascript",
							})
						}
					}
					index = valueNext
					continue
				}
			}
			index = next
		case '{', '[':
			depth++
			index++
		case '}', ']':
			depth--
			index++
		case '/':
			if next, ok := jsonCommentEnd(content, index); ok {
				index = next
			} else {
				index++
			}
		default:
			index++
		}
	}
	return result
}

func jsonObjectSingleStringArrayValueReplacements(file string, content []byte, objectRange jsonRange, rewrites map[string]string, reason string, rule string) []replacements.Replacement {
	var result []replacements.Replacement
	depth := 1
	for index := objectRange.start + 1; index < objectRange.end-1; {
		switch content[index] {
		case '"':
			_, _, _, next, ok := jsonStringToken(content, index)
			if !ok {
				return nil
			}
			if depth == 1 {
				colon := nextNonJSONWhitespace(content, next)
				valueStart := nextNonJSONWhitespace(content, colon+1)
				if colon < objectRange.end && content[colon] == ':' && valueStart < objectRange.end && content[valueStart] == '[' {
					value, rawStart, rawEnd, valueNext, ok := singleStringJSONArray(content, valueStart, objectRange.end)
					if ok {
						rawValue := string(content[rawStart:rawEnd])
						if value == rawValue {
							if replacement, ok := rewrites[value]; ok {
								result = append(result, replacements.Replacement{
									File:        file,
									Start:       rawStart,
									End:         rawEnd,
									Replacement: replacement,
									Reason:      reason,
									Rule:        rule,
									Adapter:     "javascript",
								})
							}
						}
						index = valueNext
						continue
					}
				}
			}
			index = next
		case '{', '[':
			depth++
			index++
		case '}', ']':
			depth--
			index++
		case '/':
			if next, ok := jsonCommentEnd(content, index); ok {
				index = next
			} else {
				index++
			}
		default:
			index++
		}
	}
	return result
}

func singleStringJSONArray(content []byte, arrayStart int, limit int) (string, int, int, int, bool) {
	index := nextNonJSONWhitespace(content, arrayStart+1)
	if index >= limit || index >= len(content) || content[index] != '"' {
		return "", 0, 0, 0, false
	}

	value, rawStart, rawEnd, next, ok := jsonStringToken(content, index)
	if !ok {
		return "", 0, 0, 0, false
	}
	arrayEnd := nextNonJSONWhitespace(content, next)
	if arrayEnd < limit && arrayEnd < len(content) && content[arrayEnd] == ',' {
		arrayEnd = nextNonJSONWhitespace(content, arrayEnd+1)
	}
	if arrayEnd >= limit || arrayEnd >= len(content) || content[arrayEnd] != ']' {
		return "", 0, 0, 0, false
	}
	return value, rawStart, rawEnd, arrayEnd + 1, true
}

func jsonStringToken(content []byte, quoteStart int) (string, int, int, int, bool) {
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

func nextNonJSONWhitespace(content []byte, index int) int {
	for index < len(content) {
		switch content[index] {
		case ' ', '\t', '\n', '\r':
			index++
		case '/':
			if next, ok := jsonCommentEnd(content, index); ok {
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

func jsonCommentEnd(content []byte, index int) (int, bool) {
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
