//go:build cgo

package rules_test

import (
	"testing"

	"refactorlah/internal/adapters/php"
	"refactorlah/internal/adapters/php/rules"
)

func TestFullyQualifiedClassNameRuleUpdatesExactQualifiedReference(t *testing.T) {
	source := []byte(`<?php
namespace App\Http\Controllers;

final class InvoiceController
{
    public function show(): \App\Services\Billing\InvoiceService
    {
        return new \App\Services\Billing\InvoiceService();
    }
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.FullyQualifiedClassNameRule{}.Collect(document, rules.FullyQualifiedClassNameInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %#v", replacements)
	}
	for _, replacement := range replacements {
		if string(source[replacement.Start:replacement.End]) != "\\App\\Services\\Billing\\InvoiceService" {
			t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
		}
		if replacement.Replacement != "\\App\\Domain\\Billing\\InvoiceService" {
			t.Fatalf("expected leading slash to be preserved, got %q", replacement.Replacement)
		}
	}
}

func TestFullyQualifiedClassNameRuleSkipsNamespaceAndUseDeclarations(t *testing.T) {
	source := []byte(`<?php
namespace App\Services\Billing;

use App\Services\Billing\InvoiceService;

final class InvoiceController {}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.FullyQualifiedClassNameRule{}.Collect(document, rules.FullyQualifiedClassNameInput{
		File:      "app/Http/Controllers/InvoiceController.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestFullyQualifiedClassNameRuleUpdatesClassLikeReferences(t *testing.T) {
	source := []byte(`<?php
final class HtmlRichTextRenderer implements \App\Shared\RichText\RichTextBlockRenderer {}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.FullyQualifiedClassNameRule{}.Collect(document, rules.FullyQualifiedClassNameInput{
		File:      "app/HtmlRichTextRenderer.php",
		OldSymbol: "App\\Shared\\RichText\\RichTextBlockRenderer",
		NewSymbol: "App\\Shared\\RichText\\RichTextRenderableRenderer",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %#v", replacements)
	}
	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "\\App\\Shared\\RichText\\RichTextBlockRenderer" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "\\App\\Shared\\RichText\\RichTextRenderableRenderer" {
		t.Fatalf("expected leading slash to be preserved, got %q", replacement.Replacement)
	}
}

func TestFullyQualifiedClassNameRuleDoesNotRewriteLongerSimilarReference(t *testing.T) {
	source := []byte("<?php\n$value = \\App\\Services\\Billing\\InvoiceServiceFactory::class;\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.FullyQualifiedClassNameRule{}.Collect(document, rules.FullyQualifiedClassNameInput{
		File:      "app/Factory.php",
		OldSymbol: "App\\Services\\Billing\\InvoiceService",
		NewSymbol: "App\\Domain\\Billing\\InvoiceService",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
