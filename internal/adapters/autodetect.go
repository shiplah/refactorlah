package adapters

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"refactorlah/internal/planning"
)

type Signals struct {
	PHPRelevant bool
	IncludePHP  bool
	IncludeTwig bool
}

type AutoDetector struct{}

func NewAutoDetector() *AutoDetector {
	return &AutoDetector{}
}

func (d *AutoDetector) Detect(ctx context.Context, projectRoot string, plan planning.MovePlan) (Signals, error) {
	_ = ctx

	composerPath := filepath.Join(projectRoot, "composer.json")
	composerBytes, err := os.ReadFile(composerPath)
	if err != nil {
		if os.IsNotExist(err) {
			composerPath, composerBytes, err = nestedComposerConfig(projectRoot, plan)
			if err != nil {
				return Signals{}, err
			}
			if composerPath == "" {
				return Signals{}, nil
			}
		} else {
			return Signals{}, err
		}
	}

	hasPsr4 := composerHasPSR4(composerBytes)
	includePHP := plan.ContainsExtension(".php")
	includeTwig := plan.ContainsExtension(".twig")

	if !includeTwig {
		if info, err := os.Stat(filepath.Join(projectRoot, "templates")); err == nil && info.IsDir() {
			includeTwig = true
		} else if composerPath != "" {
			composerDir := filepath.Dir(composerPath)
			if info, err := os.Stat(filepath.Join(composerDir, "templates")); err == nil && info.IsDir() {
				includeTwig = true
			}
			if info, err := os.Stat(filepath.Join(composerDir, "config", "packages", "twig.yaml")); err == nil && !info.IsDir() {
				includeTwig = true
			}
		}
	}

	return Signals{
		PHPRelevant: includePHP || includeTwig || hasPsr4,
		IncludePHP:  includePHP || hasPsr4,
		IncludeTwig: includeTwig,
	}, nil
}

func nestedComposerConfig(projectRoot string, plan planning.MovePlan) (string, []byte, error) {
	for _, move := range plan.Moves {
		candidates := []string{move.OldPath, move.NewPath}
		for _, path := range candidates {
			for _, dir := range composerCandidateDirs(path) {
				composerPath := filepath.Join(projectRoot, filepath.FromSlash(dir), "composer.json")
				if bytes, err := os.ReadFile(composerPath); err == nil {
					return composerPath, bytes, nil
				} else if !os.IsNotExist(err) {
					return "", nil, err
				}
			}
		}
	}

	return "", nil, nil
}

func composerCandidateDirs(path string) []string {
	normalized := filepath.ToSlash(path)
	if strings.Contains(filepath.Base(normalized), ".") {
		normalized = filepath.ToSlash(filepath.Dir(filepath.FromSlash(normalized)))
	}

	dirs := []string{}
	for normalized != "." && normalized != "/" && normalized != "" {
		dirs = append(dirs, normalized)
		next := filepath.ToSlash(filepath.Dir(filepath.FromSlash(normalized)))
		if next == normalized {
			break
		}
		normalized = next
	}

	return dirs
}

func composerHasPSR4(contents []byte) bool {
	var decoded struct {
		Autoload struct {
			PSR4 map[string]any `json:"psr-4"`
		} `json:"autoload"`
		AutoloadDev struct {
			PSR4 map[string]any `json:"psr-4"`
		} `json:"autoload-dev"`
	}
	if err := json.Unmarshal(contents, &decoded); err != nil {
		return false
	}
	return len(decoded.Autoload.PSR4) > 0 || len(decoded.AutoloadDev.PSR4) > 0
}
