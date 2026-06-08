package javascript

import (
	adapterproto "github.com/NickSdot/refactorlah/internal/adapters/contract"
	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
	"github.com/NickSdot/refactorlah/internal/adapters/staticimports"
	"github.com/NickSdot/refactorlah/internal/planning"
	"github.com/NickSdot/refactorlah/internal/replacements"
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
