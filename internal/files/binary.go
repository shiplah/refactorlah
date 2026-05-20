package files

func LooksBinary(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	limit := len(content)
	if limit > 8000 {
		limit = 8000
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
