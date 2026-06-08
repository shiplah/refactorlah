package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

const (
	typeScriptPathAliasReason  = "javascript-typescript-path-alias"
	typeScriptPathAliasRule    = "javascript.TypeScriptPathAliasRule"
	typeScriptPathTargetReason = "javascript-typescript-path-target"
	typeScriptPathTargetRule   = "javascript.TypeScriptPathTargetRule"
)

type typeScriptPathConfig struct {
	file     string
	content  []byte
	pathBase string
	targets  []typeScriptPathTarget
	mappings []pathAliasMapping
}

type typeScriptPathTarget struct {
	target string
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

		pathBase := typeScriptPathBase(configPath, raw.CompilerOptions.BaseURL)
		mappings, err := buildPathAliasMappings(projectRoot, pathBase, raw.CompilerOptions)
		if err != nil {
			return typeScriptPathConfig{}, true, err
		}
		return typeScriptPathConfig{
			file:     configName,
			content:  content,
			pathBase: pathBase,
			targets:  buildTypeScriptPathTargets(raw.CompilerOptions),
			mappings: mappings,
		}, true, nil
	}

	return typeScriptPathConfig{}, false, nil
}

func typeScriptPathBase(configPath string, baseURL string) string {
	configDir := filepath.Dir(configPath)
	if baseURL != "" {
		return filepath.Join(configDir, filepath.FromSlash(baseURL))
	}
	return configDir
}

func buildPathAliasMappings(projectRoot string, pathBase string, options rawTypeScriptCompilerOptions) ([]pathAliasMapping, error) {
	if len(options.Paths) == 0 {
		return nil, nil
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

func buildTypeScriptPathTargets(options rawTypeScriptCompilerOptions) []typeScriptPathTarget {
	if len(options.Paths) == 0 {
		return nil
	}

	aliasPatterns := make([]string, 0, len(options.Paths))
	for aliasPattern := range options.Paths {
		aliasPatterns = append(aliasPatterns, aliasPattern)
	}
	sort.Strings(aliasPatterns)

	var targets []typeScriptPathTarget
	for _, aliasPattern := range aliasPatterns {
		if strings.Contains(aliasPattern, "*") {
			continue
		}

		targetValues := options.Paths[aliasPattern]
		if len(targetValues) != 1 || strings.Contains(targetValues[0], "*") {
			continue
		}
		if !isJavaScriptModuleExtension(filepath.Ext(targetValues[0])) {
			continue
		}
		targets = append(targets, typeScriptPathTarget{target: targetValues[0]})
	}
	return targets
}

func pathAliasSpecifierRewrites(config typeScriptPathConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return specifierRewritesForPathAliases(config.mappings, moves, typeScriptPathAliasReason, typeScriptPathAliasRule)
}

func typeScriptPathTargetReplacements(projectRoot string, config typeScriptPathConfig, moves []planning.FileMove) []replacements.Replacement {
	targetRewrites := typeScriptPathTargetRewrites(projectRoot, config.pathBase, config.targets, moves)
	if len(targetRewrites) == 0 {
		return nil
	}

	compilerOptionsRange, ok := jsonObjectPropertyRange(config.content, "compilerOptions")
	if !ok {
		return nil
	}
	pathsRange, ok := jsonObjectPropertyRangeIn(config.content, compilerOptionsRange, "paths")
	if !ok {
		return nil
	}
	return jsonObjectSingleStringArrayValueReplacements(config.file, config.content, pathsRange, targetRewrites, typeScriptPathTargetReason, typeScriptPathTargetRule)
}

func typeScriptPathTargetRewrites(projectRoot string, pathBase string, targets []typeScriptPathTarget, moves []planning.FileMove) map[string]string {
	rewrites := map[string]string{}
	for _, target := range targets {
		for _, move := range moves {
			oldReference, ok := typeScriptTargetReference(projectRoot, pathBase, move.OldPath, target.target)
			if !ok || oldReference != target.target {
				continue
			}
			newReference, ok := typeScriptTargetReference(projectRoot, pathBase, move.NewPath, target.target)
			if !ok || oldReference == newReference {
				continue
			}
			rewrites[oldReference] = newReference
		}
	}
	return rewrites
}

func typeScriptTargetReference(projectRoot string, pathBase string, targetPath string, existingStyle string) (string, bool) {
	if !isJavaScriptModuleExtension(filepath.Ext(targetPath)) {
		return "", false
	}

	absoluteTarget := filepath.Join(projectRoot, filepath.FromSlash(targetPath))
	relative, err := filepath.Rel(pathBase, absoluteTarget)
	if err != nil {
		return "", false
	}
	relative = filepath.ToSlash(relative)
	if relative == "." || filepath.IsAbs(relative) || startsWithParentTraversal(relative) {
		return "", false
	}

	relative = strings.TrimPrefix(relative, "./")
	if strings.HasPrefix(existingStyle, "./") {
		return "./" + relative, true
	}
	return relative, true
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
