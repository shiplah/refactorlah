package rules_test

import (
	"testing"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/php/symfony/twig/rules"
)

func TestSymfonyRulesRewriteExactStaticTemplateReferences(t *testing.T) {
	tests := []struct {
		name    string
		rule    rules.StringReplacementRule
		content string
		oldText string
	}{
		{
			name:    "render",
			rule:    rules.RenderTemplateRule(),
			content: `$this->render('admin/card.html.twig', []);`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "renderView",
			rule:    rules.RenderTemplateRule(),
			content: `$this->renderView("admin/card.html.twig", []);`,
			oldText: `"admin/card.html.twig"`,
		},
		{
			name:    "template attribute",
			rule:    rules.TemplateAttributeRule(),
			content: `#[Template('admin/card.html.twig')]`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "component template attribute",
			rule:    rules.ComponentTemplateAttributeRule(),
			content: `#[AsTwigComponent(name: 'card', template: 'admin/card.html.twig')]`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "yaml template",
			rule:    rules.YamlTemplateRule(),
			content: `template: 'admin/card.html.twig'`,
			oldText: `'admin/card.html.twig'`,
		},
		{
			name:    "yaml component directory",
			rule:    rules.YamlComponentTemplateDirectoryRule(),
			content: `template_directory: 'admin'`,
			oldText: `'admin'`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapping := adapterproto.PathMapping{
				OldReference: "admin/card.html.twig",
				NewReference: "backoffice/card.html.twig",
			}
			if test.oldText == "'admin'" {
				mapping.OldReference = "admin"
				mapping.NewReference = "backoffice"
			}

			replacements := test.rule.Collect("file", test.content, mapping)
			if len(replacements) != 1 {
				t.Fatalf("expected one replacement, got %#v", replacements)
			}
			replacement := replacements[0]
			if test.content[replacement.Start:replacement.End] != test.oldText {
				t.Fatalf("replacement range points to %q", test.content[replacement.Start:replacement.End])
			}
		})
	}
}

func TestSymfonyRulesSkipDynamicRenderReferences(t *testing.T) {
	mapping := adapterproto.PathMapping{
		OldReference: "admin/card.html.twig",
		NewReference: "backoffice/card.html.twig",
	}

	replacements := rules.RenderTemplateRule().Collect(
		"Controller.php",
		`$this->render($section.'/card.html.twig', []);`,
		mapping,
	)

	if len(replacements) != 0 {
		t.Fatalf("expected no replacements, got %#v", replacements)
	}
}
