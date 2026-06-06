package syntax

import "strings"

func Parent(module string) string {
	separator := strings.LastIndex(module, ".")
	if separator < 0 {
		return ""
	}
	return module[:separator]
}

func Leaf(module string) string {
	separator := strings.LastIndex(module, ".")
	if separator < 0 {
		return module
	}
	return module[separator+1:]
}

func ResolveRelativeModule(packageName string, level int, moduleTail string) (string, bool) {
	if packageName == "" {
		return "", false
	}

	parts := strings.Split(packageName, ".")
	up := level - 1
	if up > len(parts) {
		return "", false
	}

	base := parts
	if up > 0 {
		base = parts[:len(parts)-up]
	}
	if moduleTail != "" {
		base = append(append([]string(nil), base...), strings.Split(moduleTail, ".")...)
	}

	return strings.Join(base, "."), true
}
