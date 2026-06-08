package javascript

import "refactorlah/internal/adapters/javascript/rules"

func buildTypeScriptPathAmbiguities(options rawTypeScriptCompilerOptions) []rules.TypeScriptPathAmbiguity {
	if len(options.Paths) == 0 {
		return nil
	}

	var ambiguities []rules.TypeScriptPathAmbiguity
	for _, aliasPattern := range sortedTypeScriptPathPatterns(options) {
		targets := options.Paths[aliasPattern]
		if len(targets) <= 1 {
			continue
		}
		ambiguities = append(ambiguities, rules.TypeScriptPathAmbiguity{
			Alias:   aliasPattern,
			Targets: append([]string(nil), targets...),
		})
	}
	return ambiguities
}
