package javascript

import (
	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

func packageSpecifierRewrites(config packageSpecifierConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return rules.PackageImportAliasRule{}.Collect(config.importMappings, config.selfReferenceMappings, moves)
}

func packageImportTargetReplacements(config packageSpecifierConfig, moves []planning.FileMove) []replacements.Replacement {
	return rules.PackageImportTargetRule{}.Collect(rules.PackageImportTargetInput{
		File:    "package.json",
		Content: config.content,
		Targets: config.importTargets,
		Moves:   moves,
	})
}

func packageImportWarnings(config packageSpecifierConfig, moves []planning.FileMove) []adapterproto.Warning {
	return rules.PackageImportWarningRule{}.Collect(rules.PackageImportWarningInput{
		File:               "package.json",
		ConditionalImports: config.conditionalImports,
		Moves:              moves,
	})
}
