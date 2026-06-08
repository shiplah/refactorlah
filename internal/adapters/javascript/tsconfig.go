package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	adapterproto "refactorlah/internal/adapters/contract"
	"refactorlah/internal/adapters/javascript/jsonconfig"
	"refactorlah/internal/adapters/javascript/rules"
	"refactorlah/internal/adapters/staticimports"
	"refactorlah/internal/planning"
	"refactorlah/internal/replacements"
)

type typeScriptPathConfig struct {
	file           string
	content        []byte
	pathBase       string
	targets        []rules.TypeScriptPathTarget
	ambiguousPaths []rules.TypeScriptPathAmbiguity
	mappings       []rules.PathAliasMapping
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
		if err := json.Unmarshal(jsonconfig.Normalise(content), &raw); err != nil {
			return typeScriptPathConfig{}, true, err
		}

		pathBase := typeScriptPathBase(configPath, raw.CompilerOptions.BaseURL)
		mappings, err := buildPathAliasMappings(projectRoot, pathBase, raw.CompilerOptions)
		if err != nil {
			return typeScriptPathConfig{}, true, err
		}
		return typeScriptPathConfig{
			file:           configName,
			content:        content,
			pathBase:       pathBase,
			targets:        buildTypeScriptPathTargets(raw.CompilerOptions),
			ambiguousPaths: buildTypeScriptPathAmbiguities(raw.CompilerOptions),
			mappings:       mappings,
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

func buildPathAliasMappings(projectRoot string, pathBase string, options rawTypeScriptCompilerOptions) ([]rules.PathAliasMapping, error) {
	if len(options.Paths) == 0 {
		return nil, nil
	}

	var mappings []rules.PathAliasMapping
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

func buildTypeScriptPathTargets(options rawTypeScriptCompilerOptions) []rules.TypeScriptPathTarget {
	if len(options.Paths) == 0 {
		return nil
	}

	aliasPatterns := make([]string, 0, len(options.Paths))
	for aliasPattern := range options.Paths {
		aliasPatterns = append(aliasPatterns, aliasPattern)
	}
	sort.Strings(aliasPatterns)

	var targets []rules.TypeScriptPathTarget
	for _, aliasPattern := range aliasPatterns {
		if strings.Contains(aliasPattern, "*") {
			continue
		}

		targetValues := options.Paths[aliasPattern]
		if len(targetValues) != 1 || strings.Contains(targetValues[0], "*") {
			continue
		}
		if !rules.IsJavaScriptModuleExtension(filepath.Ext(targetValues[0])) {
			continue
		}
		targets = append(targets, rules.TypeScriptPathTarget{Target: targetValues[0]})
	}
	return targets
}

func buildTypeScriptPathAmbiguities(options rawTypeScriptCompilerOptions) []rules.TypeScriptPathAmbiguity {
	if len(options.Paths) == 0 {
		return nil
	}

	aliasPatterns := make([]string, 0, len(options.Paths))
	for aliasPattern := range options.Paths {
		aliasPatterns = append(aliasPatterns, aliasPattern)
	}
	sort.Strings(aliasPatterns)

	var ambiguities []rules.TypeScriptPathAmbiguity
	for _, aliasPattern := range aliasPatterns {
		targets := options.Paths[aliasPattern]
		if len(targets) <= 1 {
			continue
		}
		ambiguities = append(ambiguities, rules.TypeScriptPathAmbiguity{
			Alias:   aliasPattern,
			Targets: append([]string(nil), targets...),
		})
	}
	return ambiguities
}

func pathAliasSpecifierRewrites(config typeScriptPathConfig, moves []planning.FileMove) []staticimports.SpecifierRewrite {
	return rules.TypeScriptPathAliasRule{}.Collect(config.mappings, moves)
}

func typeScriptPathTargetReplacements(projectRoot string, config typeScriptPathConfig, moves []planning.FileMove) []replacements.Replacement {
	return rules.TypeScriptPathTargetRule{}.Collect(rules.TypeScriptPathTargetInput{
		ProjectRoot: projectRoot,
		File:        config.file,
		Content:     config.content,
		PathBase:    config.pathBase,
		Targets:     config.targets,
		Moves:       moves,
	})
}

func typeScriptPathWarnings(projectRoot string, config typeScriptPathConfig, moves []planning.FileMove) []adapterproto.Warning {
	return rules.TypeScriptPathWarningRule{}.Collect(rules.TypeScriptPathWarningInput{
		ProjectRoot: projectRoot,
		File:        config.file,
		PathBase:    config.pathBase,
		Ambiguities: config.ambiguousPaths,
		Moves:       moves,
	})
}
