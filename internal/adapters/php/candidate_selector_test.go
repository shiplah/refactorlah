//go:build cgo

package php

import (
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
)

func TestCandidateFileSelectorBuildsImpactQuery(t *testing.T) {
	query := CandidateFileSelector{}.Query([]adapterproto.SymbolMapping{{
		OldPath:      "app/Billing/Domain/Archive/InvoiceLine.php",
		NewPath:      "app/Billing/Archive/Domain/InvoiceLine.php",
		OldSymbol:    "App\\Billing\\Domain\\Archive\\InvoiceLine",
		NewSymbol:    "App\\Billing\\Archive\\Domain\\InvoiceLine",
		OldNamespace: "App\\Billing\\Domain\\Archive",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	}})

	if len(query.Extensions) != 1 || query.Extensions[0] != ".php" {
		t.Fatalf("expected PHP extension query, got %#v", query.Extensions)
	}
	if len(query.IncludePaths) != 1 || query.IncludePaths[0] != "app/Billing/Domain/Archive/InvoiceLine.php" {
		t.Fatalf("expected moved file include, got %#v", query.IncludePaths)
	}

	expectedNeedles := []string{
		"App\\Billing\\Domain\\Archive",
		"App\\Billing\\Domain\\Archive\\InvoiceLine",
		"InvoiceLine",
	}
	for _, needle := range expectedNeedles {
		if !containsString(query.Needles, needle) {
			t.Fatalf("expected needle %q in %#v", needle, query.Needles)
		}
	}
}

func TestCandidateFileSelectorSkipsWhenThereAreNoMappings(t *testing.T) {
	query := CandidateFileSelector{}.Query(nil)
	if len(query.Extensions) != 0 || len(query.Needles) != 0 || len(query.IncludePaths) != 0 {
		t.Fatalf("expected empty query, got %#v", query)
	}
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
