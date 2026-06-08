package javascript

import (
	"path/filepath"
	"strings"

	"refactorlah/internal/adapters/javascript/rules"
)

func buildTypeScriptPathTargets(options rawTypeScriptCompilerOptions) []rules.TypeScriptPathTarget {
	if len(options.Paths) == 0 {
		return nil
	}

	var targets []rules.TypeScriptPathTarget
	for _, aliasPattern := range sortedTypeScriptPathPatterns(options) {
		if strings.Contains(aliasPattern, "*") {
			continue
		}

		targetValues := options.Paths[aliasPattern]
		if len(targetValues) != 1 || strings.Contains(targetValues[0], "*") {
			continue
		}
		if !rules.IsJavaScriptModuleExtension(filepath.Ext(targetValues[0])) {
			continue
		}
		targets = append(targets, rules.TypeScriptPathTarget{Target: targetValues[0]})
	}
	return targets
}
