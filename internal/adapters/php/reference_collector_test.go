//go:build cgo

package php

import (
	"testing"

	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
)

func TestAutoloadedFunctionConstantMappingsKeepsOnlyAutoloadedSymbols(t *testing.T) {
	mappings := []adapterproto.SymbolMapping{
		{
			Kind:      "constant",
			OldPath:   "src/Config/symbols.php",
			OldSymbol: "App\\Config\\DEFAULT_LIMIT",
			NewSymbol: "App\\Shared\\DEFAULT_LIMIT",
		},
		{
			Kind:      "function",
			OldPath:   "src/Config/helpers.php",
			OldSymbol: "App\\Config\\build_label",
			NewSymbol: "App\\Shared\\build_label",
		},
		{
			Kind:      "class",
			OldPath:   "src/Config/Reader.php",
			OldSymbol: "App\\Config\\Reader",
			NewSymbol: "App\\Shared\\Reader",
		},
	}

	filtered := autoloadedFunctionConstantMappings(mappings, map[string]bool{
		"src/Config/symbols.php": true,
		"src/Config/Reader.php":  true,
	})

	if len(filtered) != 1 {
		t.Fatalf("expected one autoloaded function/constant mapping, got %#v", filtered)
	}
	if filtered[0].OldSymbol != "App\\Config\\DEFAULT_LIMIT" {
		t.Fatalf("unexpected filtered mapping: %#v", filtered[0])
	}
}

func TestNonAutoloadedFunctionConstantMappingsKeepsOnlyNonAutoloadedSymbols(t *testing.T) {
	mappings := []adapterproto.SymbolMapping{
		{
			Kind:      "constant",
			OldPath:   "src/Config/symbols.php",
			OldSymbol: "App\\Config\\DEFAULT_LIMIT",
			NewSymbol: "App\\Shared\\DEFAULT_LIMIT",
		},
		{
			Kind:      "function",
			OldPath:   "src/Config/helpers.php",
			OldSymbol: "App\\Config\\build_label",
			NewSymbol: "App\\Shared\\build_label",
		},
		{
			Kind:      "class",
			OldPath:   "src/Config/Reader.php",
			OldSymbol: "App\\Config\\Reader",
			NewSymbol: "App\\Shared\\Reader",
		},
	}

	filtered := nonAutoloadedFunctionConstantMappings(mappings, map[string]bool{
		"src/Config/symbols.php": true,
	})

	if len(filtered) != 1 {
		t.Fatalf("expected one non-autoloaded function/constant mapping, got %#v", filtered)
	}
	if filtered[0].OldSymbol != "App\\Config\\build_label" {
		t.Fatalf("unexpected filtered mapping: %#v", filtered[0])
	}
}
