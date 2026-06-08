package jsonconfig

import "github.com/NickSdot/refactorlah/internal/replacements"

func StringValueReplacements(file string, content []byte, objectRange Range, rewrites map[string]string, reason string, rule string) []replacements.Replacement {
	var result []replacements.Replacement
	depth := 1
	for index := objectRange.start + 1; index < objectRange.end-1; {
		switch content[index] {
		case '"':
			_, _, _, next, ok := stringToken(content, index)
			if !ok {
				return nil
			}
			if depth == 1 {
				colon := nextNonWhitespace(content, next)
				valueStart := nextNonWhitespace(content, colon+1)
				if colon < objectRange.end && content[colon] == ':' && valueStart < objectRange.end && content[valueStart] == '"' {
					value, rawStart, rawEnd, valueNext, ok := stringToken(content, valueStart)
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
			if next, ok := commentEnd(content, index); ok {
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

func SingleStringArrayValueReplacements(file string, content []byte, objectRange Range, rewrites map[string]string, reason string, rule string) []replacements.Replacement {
	var result []replacements.Replacement
	depth := 1
	for index := objectRange.start + 1; index < objectRange.end-1; {
		switch content[index] {
		case '"':
			_, _, _, next, ok := stringToken(content, index)
			if !ok {
				return nil
			}
			if depth == 1 {
				colon := nextNonWhitespace(content, next)
				valueStart := nextNonWhitespace(content, colon+1)
				if colon < objectRange.end && content[colon] == ':' && valueStart < objectRange.end && content[valueStart] == '[' {
					value, rawStart, rawEnd, valueNext, ok := singleStringArray(content, valueStart, objectRange.end)
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
			if next, ok := commentEnd(content, index); ok {
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

func singleStringArray(content []byte, arrayStart int, limit int) (string, int, int, int, bool) {
	index := nextNonWhitespace(content, arrayStart+1)
	if index >= limit || index >= len(content) || content[index] != '"' {
		return "", 0, 0, 0, false
	}

	value, rawStart, rawEnd, next, ok := stringToken(content, index)
	if !ok {
		return "", 0, 0, 0, false
	}
	arrayEnd := nextNonWhitespace(content, next)
	if arrayEnd < limit && arrayEnd < len(content) && content[arrayEnd] == ',' {
		arrayEnd = nextNonWhitespace(content, arrayEnd+1)
	}
	if arrayEnd >= limit || arrayEnd >= len(content) || content[arrayEnd] != ']' {
		return "", 0, 0, 0, false
	}
	return value, rawStart, rawEnd, arrayEnd + 1, true
}
