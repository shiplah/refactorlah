package rules

const (
	TwigIncludeRuleName = "php.symfony.twig.TwigIncludeRule"
	TwigExtendsRuleName = "php.symfony.twig.TwigExtendsRule"
	TwigEmbedRuleName   = "php.symfony.twig.TwigEmbedRule"
	TwigUseRuleName     = "php.symfony.twig.TwigUseRule"
	TwigImportRuleName  = "php.symfony.twig.TwigImportRule"
	TwigFromRuleName    = "php.symfony.twig.TwigFromRule"
)

func TwigIncludeRule() PatternRule {
	return PatternRule{
		RuleName: TwigIncludeRuleName,
		Reason:   "twig-include",
		Patterns: func(quotedReference string) []string {
			reference := quotedReferencePattern(quotedReference)
			return []string{
				`{%\s*include\s+(` + reference + `)`,
				`\{\{\s*include\(\s*(` + reference + `)`,
			}
		},
	}
}

func TwigExtendsRule() PatternRule {
	return twigTagRule(TwigExtendsRuleName, "twig-extends", "extends")
}

func TwigEmbedRule() PatternRule {
	return twigTagRule(TwigEmbedRuleName, "twig-embed", "embed")
}

func TwigUseRule() PatternRule {
	return twigTagRule(TwigUseRuleName, "twig-use", "use")
}

func TwigImportRule() PatternRule {
	return twigTagRule(TwigImportRuleName, "twig-import", "import")
}

func TwigFromRule() PatternRule {
	return twigTagRule(TwigFromRuleName, "twig-from", "from")
}

func twigTagRule(ruleName string, reason string, tag string) PatternRule {
	return PatternRule{
		RuleName: ruleName,
		Reason:   reason,
		Patterns: func(quotedReference string) []string {
			return []string{`{%\s*` + tag + `\s+(` + quotedReferencePattern(quotedReference) + `)`}
		},
	}
}
