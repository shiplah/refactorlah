package rules

import "strings"

func PackageSelfReferenceMappings(packageName string) []PathAliasMapping {
	aliasPrefix, ok := packageSelfReferenceAliasPrefix(packageName)
	if !ok {
		return nil
	}
	return []PathAliasMapping{{
		AliasPrefix: aliasPrefix,
	}}
}

func packageSelfReferenceAliasPrefix(packageName string) (string, bool) {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" || strings.HasPrefix(packageName, ".") || strings.HasPrefix(packageName, "/") {
		return "", false
	}
	if strings.ContainsAny(packageName, "\\ \t\r\n") {
		return "", false
	}

	parts := strings.Split(packageName, "/")
	if strings.HasPrefix(packageName, "@") {
		if len(parts) != 2 || parts[0] == "@" || parts[1] == "" {
			return "", false
		}
		return packageName + "/", true
	}
	if len(parts) != 1 {
		return "", false
	}
	return packageName + "/", true
}
