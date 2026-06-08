package rules

import (
	"path"
	"strings"

	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

const (
	ModuleSpecifierReason   = "javascript-module-specifier"
	ModuleSpecifierRuleName = "javascript.ModuleSpecifierRule"
)

type ModuleSpecifierRule struct{}

func ModuleCandidateNeedles(moves []planning.FileMove) []string {
	seen := map[string]bool{}
	var needles []string
	for _, move := range moves {
		for _, needle := range moduleNeedlesForPath(move.OldPath) {
			if needle == "" || needle == "." || seen[needle] {
				continue
			}
			seen[needle] = true
			needles = append(needles, needle)
		}
	}
	return needles
}

func moduleNeedlesForPath(targetPath string) []string {
	needles := []string{targetPath, path.Base(targetPath)}
	extension := path.Ext(targetPath)
	if !IsJavaScriptModuleExtension(extension) {
		return needles
	}

	base := path.Base(targetPath)
	trimmedBase := strings.TrimSuffix(base, extension)
	if strings.EqualFold(trimmedBase, "index") {
		dir := path.Dir(targetPath)
		return append(needles, dir, path.Base(dir))
	}

	return append(needles, strings.TrimSuffix(targetPath, extension), trimmedBase)
}

func (r ModuleSpecifierRule) Collect(importingFile string, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	var rewrites []staticimports.SpecifierRewrite
	for _, move := range moves {
		addModuleSpecifierRewrite(&rewrites, relativeReference(importingFile, move.OldPath, false), relativeReference(importingFile, move.NewPath, false))

		oldImplicit, oldOK := implicitModuleSpecifier(importingFile, move.OldPath)
		newImplicit, newOK := implicitModuleSpecifier(importingFile, move.NewPath)
		if oldOK && newOK {
			addModuleSpecifierRewrite(&rewrites, oldImplicit, newImplicit)
		}
	}
	return rewrites
}

func addModuleSpecifierRewrite(target *[]staticimports.SpecifierRewrite, oldSpecifier string, newSpecifier string) {
	if oldSpecifier == "" || newSpecifier == "" || oldSpecifier == newSpecifier {
		return
	}
	for _, existing := range *target {
		if existing.OldSpecifier == oldSpecifier && existing.NewSpecifier == newSpecifier {
			return
		}
	}
	*target = append(*target, staticimports.SpecifierRewrite{
		OldSpecifier: oldSpecifier,
		NewSpecifier: newSpecifier,
		Reason:       ModuleSpecifierReason,
		Rule:         ModuleSpecifierRuleName,
		Adapter:      "javascript",
	})
}

func implicitModuleSpecifier(importingFile string, targetPath string) (string, bool) {
	extension := path.Ext(targetPath)
	if !IsJavaScriptModuleExtension(extension) {
		return "", false
	}

	base := path.Base(targetPath)
	if strings.EqualFold(strings.TrimSuffix(base, extension), "index") {
		specifier := relativeReference(importingFile, path.Dir(targetPath), true)
		if specifier == "." {
			return "", false
		}
		return specifier, true
	}

	return relativeReference(importingFile, strings.TrimSuffix(targetPath, extension), false), true
}

func relativeReference(importingFile string, targetPath string, targetIsDirectory bool) string {
	fromParts := pathParts(path.Dir(importingFile))
	targetParts := pathParts(targetPath)

	common := 0
	for common < len(fromParts) && common < len(targetParts) && fromParts[common] == targetParts[common] {
		common++
	}

	relativeParts := append(make([]string, 0, len(fromParts)-common+len(targetParts)-common), stringsRepeat("..", len(fromParts)-common)...)
	relativeParts = append(relativeParts, targetParts[common:]...)
	relative := strings.Join(relativeParts, "/")
	if relative == "" {
		if targetIsDirectory {
			return "."
		}
		return "./" + lastPart(targetPath)
	}
	if !strings.HasPrefix(relative, "..") {
		return "./" + relative
	}
	return relative
}

func pathParts(input string) []string {
	trimmed := strings.Trim(input, "/.")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func lastPart(input string) string {
	parts := pathParts(input)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func stringsRepeat(value string, count int) []string {
	result := make([]string, count)
	for index := range result {
		result[index] = value
	}
	return result
}
