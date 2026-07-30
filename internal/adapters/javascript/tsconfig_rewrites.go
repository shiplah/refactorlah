package javascript

import (
	adapterproto "github.com/shiplah/refactorlah/internal/adapters/contract"
	"github.com/shiplah/refactorlah/internal/adapters/javascript/rules"
	"github.com/shiplah/refactorlah/internal/adapters/staticimports"
	"github.com/shiplah/refactorlah/internal/planning"
	"github.com/shiplah/refactorlah/internal/replacements"
)

func pathAliasSpecifierRewrites(config typeScriptPathConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return rules.TypeScriptPathAliasRule{}.Collect(config.mappings, moves)
}

func typeScriptPathTargetReplacements(projectRoot string, config typeScriptPathConfig, moves []planning.FileMove) []replacements.Replacement {
	return rules.TypeScriptPathTargetRule{}.Collect(rules.TypeScriptPathTargetInput{
		ProjectRoot: projectRoot,
		File:        config.file,
		Content:     config.content,
		PathBase:    config.pathBase,
		Targets:     config.targets,
		Moves:       moves,
	})
}

func typeScriptPathWarnings(projectRoot string, config typeScriptPathConfig, moves []planning.FileMove) []adapterproto.Warning {
	return rules.TypeScriptPathWarningRule{}.Collect(rules.TypeScriptPathWarningInput{
		ProjectRoot: projectRoot,
		File:        config.file,
		PathBase:    config.pathBase,
		Ambiguities: config.ambiguousPaths,
		Moves:       moves,
	})
}
