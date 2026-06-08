package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/adapters/scan"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
)

const (
	typeScriptPathAliasReason = "javascript-typescript-path-alias"
	typeScriptPathAliasRule   = "javascript.TypeScriptPathAliasRule"
)

type typeScriptPathConfig struct {
	mappings []pathAliasMapping
}

type pathAliasMapping struct {
	aliasPrefix  string
	targetPrefix string
}

type rawTypeScriptConfig struct {
	CompilerOptions rawTypeScriptCompilerOptions `json:"compilerOptions"`
}

type rawTypeScriptCompilerOptions struct {
	BaseURL string              `json:"baseUrl"`
	Paths   map[string][]string `json:"paths"`
}

func readTypeScriptPathConfig(projectRoot string) (typeScriptPathConfig, bool, error) {
	for _, configName := range []string{"tsconfig.json", "jsconfig.json"} {
		configPath := filepath.Join(projectRoot, configName)
		content, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return typeScriptPathConfig{}, false, err
		}

		var raw rawTypeScriptConfig
		if err := json.Unmarshal(normaliseJSONConfig(content), &raw); err != nil {
			return typeScriptPathConfig{}, true, err
		}

		mappings, err := buildPathAliasMappings(projectRoot, configPath, raw.CompilerOptions)
		if err != nil {
			return typeScriptPathConfig{}, true, err
		}
		return typeScriptPathConfig{mappings: mappings}, true, nil
	}

	return typeScriptPathConfig{}, false, nil
}

func buildPathAliasMappings(projectRoot string, configPath string, options rawTypeScriptCompilerOptions) ([]pathAliasMapping, error) {
	if len(options.Paths) == 0 {
		return nil, nil
	}

	configDir := filepath.Dir(configPath)
	pathBase := configDir
	if options.BaseURL != "" {
		pathBase = filepath.Join(configDir, filepath.FromSlash(options.BaseURL))
	}

	var mappings []pathAliasMapping
	aliasPatterns := make([]string, 0, len(options.Paths))
	for aliasPattern := range options.Paths {
		aliasPatterns = append(aliasPatterns, aliasPattern)
	}
	sort.Strings(aliasPatterns)

	for _, aliasPattern := range aliasPatterns {
		targets := options.Paths[aliasPattern]
		if len(targets) != 1 {
			continue
		}

		aliasPrefix, aliasOK := wildcardPrefix(aliasPattern)
		targetPrefix, targetOK := wildcardPrefix(targets[0])
		if !aliasOK || !targetOK {
			continue
		}

		resolvedPrefix, ok, err := resolveAliasTargetPrefix(projectRoot, pathBase, targetPrefix)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		mappings = append(mappings, pathAliasMapping{
			aliasPrefix:  aliasPrefix,
			targetPrefix: resolvedPrefix,
		})
	}

	return mappings, nil
}

func resolveAliasTargetPrefix(projectRoot string, pathBase string, targetPrefix string) (string, bool, error) {
	resolved := filepath.Clean(filepath.Join(pathBase, filepath.FromSlash(targetPrefix)))
	relative, err := filepath.Rel(projectRoot, resolved)
	if err != nil {
		return "", false, err
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || filepath.IsAbs(relative) || startsWithParentTraversal(relative) {
		return "", false, nil
	}

	if relative == "." {
		return "", true, nil
	}
	return strings.TrimSuffix(relative, "/") + "/", true, nil
}

func wildcardPrefix(pattern string) (string, bool) {
	if strings.Count(pattern, "*") != 1 || !strings.HasSuffix(pattern, "*") {
		return "", false
	}
	return strings.TrimSuffix(pattern, "*"), true
}

func pathAliasSpecifierRewrites(config typeScriptPathConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	rewrites := map[string]string{}
	conflicts := map[string]bool{}

	for _, mapping := range config.mappings {
		for _, move := range moves {
			oldSuffix, oldOK := moduleSpecifierWithinTarget(move.OldPath, mapping.targetPrefix)
			newSuffix, newOK := moduleSpecifierWithinTarget(move.NewPath, mapping.targetPrefix)
			if !oldOK || !newOK {
				continue
			}

			oldSpecifier := mapping.aliasPrefix + oldSuffix
			newSpecifier := mapping.aliasPrefix + newSuffix
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
			Reason:       typeScriptPathAliasReason,
			Rule:         typeScriptPathAliasRule,
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
	if !isJavaScriptModuleExtension(extension) {
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

func typeScriptAliasCandidateQuery(rewrites []staticimports.SpecifierRewrite) scan.CandidateQuery {
	query := scan.CandidateQuery{
		Extensions: []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"},
	}
	for _, rewrite := range rewrites {
		query.Needles = append(query.Needles, rewrite.OldSpecifier)
	}
	return query
}

func startsWithParentTraversal(path string) bool {
	return len(path) > 3 && path[:3] == "../"
}

func isJavaScriptModuleExtension(extension string) bool {
	switch extension {
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}

func normaliseJSONConfig(content []byte) []byte {
	return removeTrailingJSONCommas(stripJSONComments(content))
}

func stripJSONComments(content []byte) []byte {
	result := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for index := 0; index < len(content); index++ {
		current := content[index]
		if inString {
			result = append(result, current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}

		if current == '"' {
			inString = true
			result = append(result, current)
			continue
		}
		if current == '/' && index+1 < len(content) && content[index+1] == '/' {
			index += 2
			for index < len(content) && content[index] != '\n' {
				index++
			}
			if index < len(content) {
				result = append(result, content[index])
			}
			continue
		}
		if current == '/' && index+1 < len(content) && content[index+1] == '*' {
			index += 2
			for index+1 < len(content) && !(content[index] == '*' && content[index+1] == '/') {
				if content[index] == '\n' {
					result = append(result, '\n')
				}
				index++
			}
			index++
			continue
		}
		result = append(result, current)
	}

	return result
}

func removeTrailingJSONCommas(content []byte) []byte {
	result := make([]byte, 0, len(content))
	inString := false
	escaped := false

	for index := 0; index < len(content); index++ {
		current := content[index]
		if inString {
			result = append(result, current)
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inString = false
			}
			continue
		}

		if current == '"' {
			inString = true
			result = append(result, current)
			continue
		}
		if current == ',' && nextNonWhitespace(content, index+1) >= 0 {
			next := nextNonWhitespace(content, index+1)
			if content[next] == '}' || content[next] == ']' {
				continue
			}
		}
		result = append(result, current)
	}

	return result
}

func nextNonWhitespace(content []byte, index int) int {
	for index < len(content) {
		switch content[index] {
		case ' ', '\t', '\n', '\r':
			index++
		default:
			return index
		}
	}
	return -1
}
