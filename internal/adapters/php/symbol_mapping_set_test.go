//go:build cgo

package php

import (
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
)

func TestSymbolMappingSetIndexesMappingsAndReturnsDefensiveCopies(t *testing.T) {
	mapping := adapterproto.SymbolMapping{
		OldPath:   "app/Old/InvoiceService.php",
		NewPath:   "app/New/InvoiceProcessor.php",
		OldSymbol: "App\\Old\\InvoiceService",
		NewSymbol: "App\\New\\InvoiceProcessor",
	}
	set := NewSymbolMappingSet([]adapterproto.SymbolMapping{mapping})

	if set.Len() != 1 {
		t.Fatalf("expected one mapping, got %d", set.Len())
	}
	movedMapping, ok := set.MovedFile("app/Old/InvoiceService.php")
	if !ok {
		t.Fatal("expected moved file lookup")
	}
	if movedMapping.NewSymbol != "App\\New\\InvoiceProcessor" {
		t.Fatalf("unexpected moved mapping: %#v", movedMapping)
	}

	references := set.References()
	if len(references) != 1 || references[0].OldSymbol != "App\\Old\\InvoiceService" || references[0].NewSymbol != "App\\New\\InvoiceProcessor" {
		t.Fatalf("unexpected references: %#v", references)
	}

	allMappings := set.All()
	allMappings[0].NewSymbol = "Mutated"
	references[0].NewSymbol = "Mutated"

	movedMapping, _ = set.MovedFile("app/Old/InvoiceService.php")
	if movedMapping.NewSymbol != "App\\New\\InvoiceProcessor" {
		t.Fatalf("expected mapping set to keep defensive copy, got %#v", movedMapping)
	}
	if set.References()[0].NewSymbol != "App\\New\\InvoiceProcessor" {
		t.Fatalf("expected references to be defensively copied, got %#v", set.References())
	}
}
