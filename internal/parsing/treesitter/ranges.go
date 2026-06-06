//go:build cgo

package treesitter

func NodeInsideAnyRange(node Node, ranges []Node) bool {
	for _, candidate := range ranges {
		if node.StartByte >= candidate.StartByte && node.EndByte <= candidate.EndByte {
			return true
		}
	}
	return false
}
