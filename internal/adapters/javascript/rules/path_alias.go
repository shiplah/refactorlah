package rules

import (
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

type PathAliasMapping struct {
	AliasPrefix  string
	TargetPrefix string
}

type PathAliasSpecifierRule struct {
	Reason string
	Rule   string
}

func ResolveAliasTargetPrefix(projectRoot string, pathBase string, targetPrefix string) (string, bool, error) {
	resolved := filepath.Clean(filepath.Join(pathBase, filepath.FromSlash(targetPrefix)))
	relative, err := filepath.Rel(projectRoot, resolved)
	if err != nil {
		return "", false, err
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || filepath.IsAbs(relative) || StartsWithParentTraversal(relative) {
		return "", false, nil
	}

	if relative == "." {
		return "", true, nil
	}
	return strings.TrimSuffix(relative, "/") + "/", true, nil
}

func WildcardPrefix(pattern string) (string, bool) {
	if strings.Count(pattern, "*") != 1 || !strings.HasSuffix(pattern, "*") {
		return "", false
	}
	return strings.TrimSuffix(pattern, "*"), true
}

func (r PathAliasSpecifierRule) Collect(mappings []PathAliasMapping, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := map[string]string{}
	conflicts := map[string]bool{}

	for _, mapping := range mappings {
		for _, move := range moves {
			oldSuffix, oldOK := moduleSpecifierWithinTarget(move.OldPath, mapping.TargetPrefix)
			newSuffix, newOK := moduleSpecifierWithinTarget(move.NewPath, mapping.TargetPrefix)
			if !oldOK || !newOK {
				continue
			}

			oldSpecifier := mapping.AliasPrefix + oldSuffix
			newSpecifier := mapping.AliasPrefix + newSuffix
			if oldSpecifier == newSpecifier {
				continue
			}

			if existing, ok := rewrites[oldSpecifier]; ok && existing != newSpecifier {
				conflicts[oldSpecifier] = true
				delete(rewrites, oldSpecifier)
				continue
			}
			if conflicts[oldSpecifier] {
				continue
			}
			rewrites[oldSpecifier] = newSpecifier
		}
	}

	result := make([]staticimports.SpecifierRewrite, 0, len(rewrites))
	oldSpecifiers := make([]string, 0, len(rewrites))
	for oldSpecifier := range rewrites {
		oldSpecifiers = append(oldSpecifiers, oldSpecifier)
	}
	sort.Strings(oldSpecifiers)

	for _, oldSpecifier := range oldSpecifiers {
		result = append(result, staticimports.SpecifierRewrite{
			OldSpecifier: oldSpecifier,
			NewSpecifier: rewrites[oldSpecifier],
			Reason:       r.Reason,
			Rule:         r.Rule,
			Adapter:      "javascript",
		})
	}
	return result
}

func moduleSpecifierWithinTarget(targetPath string, targetPrefix string) (string, bool) {
	if targetPrefix == "" {
		return implicitModulePath(targetPath)
	}
	if !strings.HasPrefix(targetPath, targetPrefix) {
		return "", false
	}
	return implicitModulePath(strings.TrimPrefix(targetPath, targetPrefix))
}

func implicitModulePath(targetPath string) (string, bool) {
	extension := filepath.Ext(targetPath)
	if !IsJavaScriptModuleExtension(extension) {
		return "", false
	}

	targetPath = filepath.ToSlash(targetPath)
	withoutExtension := strings.TrimSuffix(targetPath, extension)
	if strings.HasSuffix(withoutExtension, "/index") {
		withoutExtension = strings.TrimSuffix(withoutExtension, "/index")
	}
	if withoutExtension == "" || withoutExtension == "." {
		return "", false
	}
	return strings.TrimPrefix(withoutExtension, "./"), true
}

func SpecifierRewriteCandidateQuery(rewrites []staticimports.SpecifierRewrite) scan.CandidateQuery {
	query := scan.CandidateQuery{
		Extensions: JavaScriptModuleExtensions(),
	}
	for _, rewrite := range rewrites {
		query.Needles = append(query.Needles, rewrite.OldSpecifier)
	}
	return query
}

func StartsWithParentTraversal(path string) bool {
	return len(path) > 3 && path[:3] == "../"
}

func IsJavaScriptModuleExtension(extension string) bool {
	switch extension {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func JavaScriptModuleExtensions() []string {
	return []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"}
}
