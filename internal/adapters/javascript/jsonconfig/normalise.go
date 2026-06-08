package jsonconfig

func Normalise(content []byte) []byte {
	return removeTrailingCommas(stripComments(content))
}

func stripComments(content []byte) []byte {
	result := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for index := 0; index < len(content); index++ {
		current := content[index]
		if inString {
			result = append(result, current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}

		if current == '"' {
			inString = true
			result = append(result, current)
			continue
		}
		if current == '/' && index+1 < len(content) && content[index+1] == '/' {
			index += 2
			for index < len(content) && content[index] != '\n' {
				index++
			}
			if index < len(content) {
				result = append(result, content[index])
			}
			continue
		}
		if current == '/' && index+1 < len(content) && content[index+1] == '*' {
			index += 2
			for index+1 < len(content) && !(content[index] == '*' && content[index+1] == '/') {
				if content[index] == '\n' {
					result = append(result, '\n')
				}
				index++
			}
			index++
			continue
		}
		result = append(result, current)
	}

	return result
}

func removeTrailingCommas(content []byte) []byte {
	result := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for index := 0; index < len(content); index++ {
		current := content[index]
		if inString {
			result = append(result, current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}

		if current == '"' {
			inString = true
			result = append(result, current)
			continue
		}
		if current == ',' {
			next := nextPlainNonWhitespace(content, index+1)
			if next < len(content) && (content[next] == '}' || content[next] == ']') {
				continue
			}
		}
		result = append(result, current)
	}

	return result
}

func nextPlainNonWhitespace(content []byte, index int) int {
	for index < len(content) {
		switch content[index] {
		case ' ', '\t', '\n', '\r':
			index++
		default:
			return index
		}
	}
	return len(content)
}
