package javascript

import (
	"encoding/json"
	"os"
	"path/filepath"

	"refactorlah/internal/adapters/javascript/jsonconfig"
	"refactorlah/internal/adapters/javascript/rules"
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
