package rules

const (
	RenderTemplateRuleName                 = "php.symfony.twig.RenderTemplateRule"
	TemplateAttributeRuleName              = "php.symfony.twig.TemplateAttributeRule"
	ComponentTemplateAttributeRuleName     = "php.symfony.twig.ComponentTemplateAttributeRule"
	YamlTemplateRuleName                   = "php.symfony.twig.YamlTemplateRule"
	YamlComponentTemplateDirectoryRuleName = "php.symfony.twig.YamlComponentTemplateDirectoryRule"
)

func RenderTemplateRule() PatternRule {
	return PatternRule{
		RuleName: RenderTemplateRuleName,
		Reason:   "symfony-render-template",
		Patterns: func(quotedReference string) []string {
			return []string{`->render(?:View)?\(\s*(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}

func TemplateAttributeRule() PatternRule {
	return PatternRule{
		RuleName: TemplateAttributeRuleName,
		Reason:   "symfony-template-attribute",
		Patterns: func(quotedReference string) []string {
			return []string{`#\[\s*Template\(\s*(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}

func ComponentTemplateAttributeRule() PatternRule {
	return PatternRule{
		RuleName: ComponentTemplateAttributeRuleName,
		Reason:   "symfony-component-template-attribute",
		Patterns: func(quotedReference string) []string {
			return []string{`#\[[^\]]*\bAsTwigComponent\b[^\]]*\btemplate\s*:\s*(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}

func YamlTemplateRule() PatternRule {
	return PatternRule{
		RuleName: YamlTemplateRuleName,
		Reason:   "yaml-template",
		Patterns: func(quotedReference string) []string {
			return []string{`\btemplate:\s*(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}

func YamlComponentTemplateDirectoryRule() PatternRule {
	return PatternRule{
		RuleName: YamlComponentTemplateDirectoryRuleName,
		Reason:   "yaml-component-template-directory",
		Patterns: func(quotedReference string) []string {
			return []string{`\btemplate_directory:\s*(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}
