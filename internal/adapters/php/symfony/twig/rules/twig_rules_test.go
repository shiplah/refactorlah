package rules_test

import (
	"testing"

	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/php/symfony/twig/rules"
)

func TestTwigRulesRewriteExactStaticReferences(t *testing.T) {
	tests := []struct {
		name    string
		rule    rules.StringReplacementRule
		content string
		oldText string
	}{
		{
			name:    "include tag",
			rule:    rules.TwigIncludeRule(),
			content: `{% include 'admin/card.html.twig' %}`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "include function",
			rule:    rules.TwigIncludeRule(),
			content: `{{ include("admin/card.html.twig") }}`,
			oldText: `"admin/card.html.twig"`,
		},
		{
			name:    "extends",
			rule:    rules.TwigExtendsRule(),
			content: `{% extends 'admin/card.html.twig' %}`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "embed",
			rule:    rules.TwigEmbedRule(),
			content: `{% embed 'admin/card.html.twig' %}`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "use",
			rule:    rules.TwigUseRule(),
			content: `{% use 'admin/card.html.twig' %}`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "import",
			rule:    rules.TwigImportRule(),
			content: `{% import 'admin/card.html.twig' as card %}`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "from",
			rule:    rules.TwigFromRule(),
			content: `{% from 'admin/card.html.twig' import card %}`,
			oldText: `'admin/card.html.twig'`,
		},
	}

	mapping := adapterproto.PathMapping{
		OldReference: "admin/card.html.twig",
		NewReference: "backoffice/card.html.twig",
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			replacements := test.rule.Collect("templates/page.html.twig", test.content, mapping)
			if len(replacements) != 1 {
				t.Fatalf("expected one replacement, got %#v", replacements)
			}
			replacement := replacements[0]
			if test.content[replacement.Start:replacement.End] != test.oldText {
				t.Fatalf("replacement range points to %q", test.content[replacement.Start:replacement.End])
			}
			if replacement.Replacement[1:len(replacement.Replacement)-1] != "backoffice/card.html.twig" {
				t.Fatalf("unexpected replacement %q", replacement.Replacement)
			}
		})
	}
}

func TestTwigRulesSkipDynamicReferences(t *testing.T) {
	mapping := adapterproto.PathMapping{
		OldReference: "admin/card.html.twig",
		NewReference: "backoffice/card.html.twig",
	}

	replacements := rules.TwigIncludeRule().Collect(
		"templates/page.html.twig",
		`{% include template_name %} {{ include(section ~ '/card.html.twig') }}`,
		mapping,
	)

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
