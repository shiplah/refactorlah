package php

import "testing"

func TestDeclarationNameOffsetFindsPrimarySymbolName(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "class",
			text: "final readonly class InvoiceBatch",
			want: "InvoiceBatch",
		},
		{
			name: "interface",
			text: "interface RichTextBlockWebRenderer extends Renderer",
			want: "RichTextBlockWebRenderer",
		},
		{
			name: "trait",
			text: "trait ComparesDocuments",
			want: "ComparesDocuments",
		},
		{
			name: "enum",
			text: "enum RichTextComponentKind: string",
			want: "RichTextComponentKind",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match, ok := DeclarationNameOffset(test.text)
			if !ok {
				t.Fatalf("expected declaration name")
			}
			if match.Name != test.want {
				t.Fatalf("expected %q, got %q", test.want, match.Name)
			}
			if test.text[match.Start:match.End] != test.want {
				t.Fatalf("offset points to %q", test.text[match.Start:match.End])
			}
		})
	}
}

func TestDeclarationNameOffsetUsesKeywordBoundaries(t *testing.T) {
	match, ok := DeclarationNameOffset("final classNameNotAKeyword {}")
	if ok {
		t.Fatalf("expected no declaration name, got %#v", match)
	}
}
