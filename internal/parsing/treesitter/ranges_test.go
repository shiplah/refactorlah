//go:build cgo

package treesitter

import "testing"

func TestNodeInsideAnyRange(t *testing.T) {
	t.Parallel()

	ranges := []Node{
		{StartByte: 10, EndByte: 20},
		{StartByte: 30, EndByte: 40},
	}

	if !NodeInsideAnyRange(Node{StartByte: 12, EndByte: 18}, ranges) {
		t.Fatal("expected node to be inside range")
	}
	if NodeInsideAnyRange(Node{StartByte: 18, EndByte: 22}, ranges) {
		t.Fatal("did not expect overlapping node to count as inside")
	}
	if NodeInsideAnyRange(Node{StartByte: 21, EndByte: 29}, ranges) {
		t.Fatal("did not expect outside node to count as inside")
	}
}
