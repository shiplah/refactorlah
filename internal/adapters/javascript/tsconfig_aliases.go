package javascript

import (
	"sort"

	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
)

func buildPathAliasMappings(projectRoot string, pathBase string, options rawTypeScriptCompilerOptions) ([]rules.PathAliasMapping, error) {
	if len(options.Paths) == 0 {
		return nil, nil
	}

	var mappings []rules.PathAliasMapping
	aliasPatterns := sortedTypeScriptPathPatterns(options)

	for _, aliasPattern := range aliasPatterns {
		targets := options.Paths[aliasPattern]
		if len(targets) != 1 {
			continue
		}

		aliasPrefix, aliasOK := rules.WildcardPrefix(aliasPattern)
		targetPrefix, targetOK := rules.WildcardPrefix(targets[0])
		if !aliasOK || !targetOK {
			continue
		}

		resolvedPrefix, ok, err := rules.ResolveAliasTargetPrefix(projectRoot, pathBase, targetPrefix)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		mappings = append(mappings, rules.PathAliasMapping{
			AliasPrefix:  aliasPrefix,
			TargetPrefix: resolvedPrefix,
		})
	}

	return mappings, nil
}

func sortedTypeScriptPathPatterns(options rawTypeScriptCompilerOptions) []string {
	aliasPatterns := make([]string, 0, len(options.Paths))
	for aliasPattern := range options.Paths {
		aliasPatterns = append(aliasPatterns, aliasPattern)
	}
	sort.Strings(aliasPatterns)
	return aliasPatterns
}
