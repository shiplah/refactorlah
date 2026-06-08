//go:build cgo

package rules_test

import (
	"testing"

	"github.com/NickSdot/refactorlah/internal/adapters/php"
	"github.com/NickSdot/refactorlah/internal/adapters/php/rules"
)

func TestClassDeclarationRuleRenamesMovedClassDeclaration(t *testing.T) {
	source := []byte("<?php\nnamespace App\\Billing\\Domain;\nfinal readonly class InvoiceIndex implements Registry {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
		File:         "app/Billing/Invoice/Domain/InvoiceIndex.php",
		OldShortName: "InvoiceIndex",
		NewShortName: "InvoiceLookup",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}

	replacement := replacements[0]
	if string(source[replacement.Start:replacement.End]) != "InvoiceIndex" {
		t.Fatalf("replacement range points to %q", string(source[replacement.Start:replacement.End]))
	}
	if replacement.Replacement != "InvoiceLookup" {
		t.Fatalf("expected replacement class name, got %q", replacement.Replacement)
	}
	if replacement.Rule != rules.ClassDeclarationRuleName {
		t.Fatalf("expected rule name %q, got %q", rules.ClassDeclarationRuleName, replacement.Rule)
	}
}

func TestClassDeclarationRuleKeepsByteOffsetsStableAfterUnicodeText(t *testing.T) {
	source := []byte("<?php\n// café before the class must not shift byte offsets.\nfinal class RichTextBlockWebRenderer {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
		File:         "app/RichTextBlockWebRenderer.php",
		OldShortName: "RichTextBlockWebRenderer",
		NewShortName: "RichTextRenderableWebRenderer",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "RichTextBlockWebRenderer" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
}

func TestClassDeclarationRuleRenamesInterfacesTraitsAndEnums(t *testing.T) {
	tests := []struct {
		name string
		text string
		old  string
		new  string
	}{
		{
			name: "interface",
			text: "<?php\ninterface RichTextBlockWebRenderer {}\n",
			old:  "RichTextBlockWebRenderer",
			new:  "RichTextRenderableWebRenderer",
		},
		{
			name: "trait",
			text: "<?php\ntrait ComparesOldDocuments {}\n",
			old:  "ComparesOldDocuments",
			new:  "ComparesDocuments",
		},
		{
			name: "enum",
			text: "<?php\nenum RichTextComponentKind: string { case Accordion = 'accordion'; }\n",
			old:  "RichTextComponentKind",
			new:  "RichTextDirectiveKind",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(test.text)
			document, err := php.Parse(source)
			if err != nil {
				t.Fatalf("parse php: %v", err)
			}
			defer document.Close()

			replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
				File:         "app/Symbol.php",
				OldShortName: test.old,
				NewShortName: test.new,
			})

			if len(replacements) != 1 {
				t.Fatalf("expected 1 replacement, got %d", len(replacements))
			}
			if string(source[replacements[0].Start:replacements[0].End]) != test.old {
				t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
			}
		})
	}
}

func TestClassDeclarationRuleDoesNotRewriteImplementedInterface(t *testing.T) {
	source := []byte("<?php\nfinal class HtmlRichTextRenderer implements RichTextBlockWebRenderer {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
		File:         "app/HtmlRichTextRenderer.php",
		OldShortName: "RichTextBlockWebRenderer",
		NewShortName: "RichTextRenderableWebRenderer",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}

func TestClassDeclarationRuleDoesNotRewriteNestedDeclaration(t *testing.T) {
	source := []byte(`<?php
namespace App\Billing\Domain;

final class InvoiceIndex {}

function buildLocalInvoiceIndex(): object
{
    class InvoiceIndex {}

    return new InvoiceIndex();
}
`)
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
		File:         "app/Billing/Domain/InvoiceIndex.php",
		OldShortName: "InvoiceIndex",
		NewShortName: "InvoiceLookup",
	})

	if len(replacements) != 1 {
		t.Fatalf("expected only the top-level declaration replacement, got %#v", replacements)
	}
	if string(source[replacements[0].Start:replacements[0].End]) != "InvoiceIndex" {
		t.Fatalf("replacement range points to %q", string(source[replacements[0].Start:replacements[0].End]))
	}
}

func TestClassDeclarationRuleDoesNotRewriteLongerSimilarDeclaration(t *testing.T) {
	source := []byte("<?php\nfinal readonly class CacheInvoiceIndex implements InvoiceIndex {}\n")
	document, err := php.Parse(source)
	if err != nil {
		t.Fatalf("parse php: %v", err)
	}
	defer document.Close()

	replacements := rules.ClassDeclarationRule{}.Collect(document, rules.ClassDeclarationInput{
		File:         "app/CacheInvoiceIndex.php",
		OldShortName: "InvoiceIndex",
		NewShortName: "InvoiceLookup",
	})

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
