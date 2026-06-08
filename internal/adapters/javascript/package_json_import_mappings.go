package javascript

import (
	"encoding/json"
	"sort"
	"strings"

	"refactorlah/internal/adapters/javascript/rules"
)

func buildPackageImportMappings(projectRoot string, imports map[string]json.RawMessage) ([]rules.PathAliasMapping, error) {
	if len(imports) == 0 {
		return nil, nil
	}

	var mappings []rules.PathAliasMapping
	for _, importKey := range sortedPackageImportKeys(imports) {
		if !strings.HasPrefix(importKey, "#") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err != nil {
			continue
		}
		if !strings.HasPrefix(target, "./") {
			continue
		}

		aliasPrefix, aliasOK := rules.WildcardPrefix(importKey)
		targetPrefix, targetOK := rules.WildcardPrefix(target)
		if !aliasOK || !targetOK {
			continue
		}

		resolvedPrefix, ok, err := rules.ResolveAliasTargetPrefix(projectRoot, projectRoot, targetPrefix)
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

func sortedPackageImportKeys(imports map[string]json.RawMessage) []string {
	importKeys := make([]string, 0, len(imports))
	for importKey := range imports {
		importKeys = append(importKeys, importKey)
	}
	sort.Strings(importKeys)
	return importKeys
}
