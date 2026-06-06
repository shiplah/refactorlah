package shared

import "testing"

func TestUpperFirst(t *testing.T) {
	t.Parallel()

	if got := UpperFirst("componentRenderer"); got != "ComponentRenderer" {
		t.Fatalf("expected ComponentRenderer, got %q", got)
	}
	if got := UpperFirst("üser"); got != "Üser" {
		t.Fatalf("expected Üser, got %q", got)
	}
	if got := UpperFirst(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestLowerFirst(t *testing.T) {
	t.Parallel()

	if got := LowerFirst("ComponentRenderer"); got != "componentRenderer" {
		t.Fatalf("expected componentRenderer, got %q", got)
	}
	if got := LowerFirst("Üser"); got != "üser" {
		t.Fatalf("expected üser, got %q", got)
	}
	if got := LowerFirst(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
