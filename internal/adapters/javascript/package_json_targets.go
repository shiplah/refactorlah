package javascript

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"refactorlah/internal/adapters/javascript/rules"
)

func buildPackageImportTargets(imports map[string]json.RawMessage) []rules.PackageImportTarget {
	if len(imports) == 0 {
		return nil
	}

	var targets []rules.PackageImportTarget
	for _, importKey := range sortedPackageImportKeys(imports) {
		if !strings.HasPrefix(importKey, "#") || strings.Contains(importKey, "*") {
			continue
		}

		var target string
		if err := json.Unmarshal(imports[importKey], &target); err != nil {
			continue
		}
		if !strings.HasPrefix(target, "./") || strings.Contains(target, "*") {
			continue
		}
		if !rules.IsJavaScriptModuleExtension(filepath.Ext(target)) {
			continue
		}

		targets = append(targets, rules.PackageImportTarget{Target: target})
	}
	return targets
}
