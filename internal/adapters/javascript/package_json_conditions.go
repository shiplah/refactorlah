package javascript

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/NickSdot/refactorlah/internal/adapters/javascript/rules"
)

func buildPackageConditionalImports(imports map[string]json.RawMessage) []rules.PackageConditionalImport {
	if len(imports) == 0 {
		return nil
	}

	var conditionalImports []rules.PackageConditionalImport
	for _, importKey := range sortedPackageImportKeys(imports) {
		if !strings.HasPrefix(importKey, "#") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err == nil {
			continue
		}

		targets := packageTargetStrings(imports[importKey])
		if len(targets) == 0 {
			continue
		}
		conditionalImports = append(conditionalImports, rules.PackageConditionalImport{
			Key:     importKey,
			Targets: targets,
		})
	}
	return conditionalImports
}

func packageTargetStrings(raw json.RawMessage) []string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var targets []string
	collectPackageTargetStrings(value, seen, &targets)
	sort.Strings(targets)
	return targets
}

func collectPackageTargetStrings(value any, seen map[string]bool, targets *[]string) {
	switch typed := value.(type) {
	case string:
		if !strings.HasPrefix(typed, "./") || seen[typed] {
			return
		}
		seen[typed] = true
		*targets = append(*targets, typed)
	case []any:
		for _, item := range typed {
			collectPackageTargetStrings(item, seen, targets)
		}
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectPackageTargetStrings(typed[key], seen, targets)
		}
	}
}
