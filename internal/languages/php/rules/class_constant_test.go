//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/languages/php"
	"refactorlah/internal/languages/php/rules"
)

func TestClassConstantRuleUpdatesClassConstantReference(t *testing.T) {
	source := []byte("<?php\n$value = \\App\\Services\\Billing\\InvoiceService::class;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassConstantRule{}.Collect(document, rules.ClassConstantInput{
		File:      "app/Config.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "\\App\\Services\\Billing\\InvoiceService" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "\\App\\Domain\\Billing\\InvoiceService" {
		t.Fatalf("expected leading slash to be preserved, got %q", replacement.Replacement)
	}
}

func TestClassConstantRuleSkipsOtherConstants(t *testing.T) {
	source := []byte("<?php\n$value = \\App\\Services\\Billing\\InvoiceService::NAME;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassConstantRule{}.Collect(document, rules.ClassConstantInput{
		File:      "app/Config.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestClassConstantRuleDoesNotRewriteLongerSimilarReference(t *testing.T) {
	source := []byte("<?php\n$value = \\App\\Services\\Billing\\InvoiceServiceFactory::class;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassConstantRule{}.Collect(document, rules.ClassConstantInput{
		File:      "app/Config.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
