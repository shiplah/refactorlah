//go:build cgo

package php

import (
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
)

func TestCandidateFileSelectorKeepsMovedFilesAndFilesMentioningMappings(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "app/Billing/Domain/Archive/InvoiceLine.php", "<?php\nnamespace App\\Billing\\Domain\\Archive;\nfinal class InvoiceLine {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Consumer/UsesInvoiceLine.php", "<?php\nnamespace App\\Consumer;\nuse App\\Billing\\Domain\\Archive\\InvoiceLine;\nfinal class UsesInvoiceLine {}\n")
	writeAnalyzerFixtureFile(t, root, "app/Consumer/UnrelatedFile.php", "<?php\nnamespace App\\Consumer;\nfinal class UnrelatedFile {}\n")

	selected := CandidateFileSelector{}.Select(root, []string{
		"app/Consumer/UnrelatedFile.php",
		"app/Consumer/UsesInvoiceLine.php",
		"app/Billing/Domain/Archive/InvoiceLine.php",
	}, []adapterproto.SymbolMapping{{
		OldPath:      "app/Billing/Domain/Archive/InvoiceLine.php",
		NewPath:      "app/Billing/Archive/Domain/InvoiceLine.php",
		OldSymbol:    "App\\Billing\\Domain\\Archive\\InvoiceLine",
		NewSymbol:    "App\\Billing\\Archive\\Domain\\InvoiceLine",
		OldNamespace: "App\\Billing\\Domain\\Archive",
		NewNamespace: "App\\Billing\\Archive\\Domain",
	}})

	expected := []string{
		"app/Consumer/UsesInvoiceLine.php",
		"app/Billing/Domain/Archive/InvoiceLine.php",
	}
	if len(selected) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, selected)
	}
	for index, file := range expected {
		if selected[index] != file {
			t.Fatalf("expected %#v, got %#v", expected, selected)
		}
	}
}

func TestCandidateFileSelectorSkipsWhenThereAreNoMappings(t *testing.T) {
	root := t.TempDir()
	writeAnalyzerFixtureFile(t, root, "app/Anything.php", "<?php\n")

	selected := CandidateFileSelector{}.Select(root, []string{"app/Anything.php"}, nil)
	if len(selected) != 0 {
		t.Fatalf("expected no selected files, got %#v", selected)
	}
}
