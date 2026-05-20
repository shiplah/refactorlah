package adapters

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

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
			return Signals{}, nil
		}
		return Signals{}, err
	}

	hasPsr4 := composerHasPSR4(composerBytes)
	includePHP := plan.ContainsExtension(".php")
	includeTwig := plan.ContainsExtension(".twig")

	if !includeTwig {
		if info, err := os.Stat(filepath.Join(projectRoot, "templates")); err == nil && info.IsDir() {
			includeTwig = true
		}
	}

	return Signals{
		PHPRelevant: includePHP || includeTwig || hasPsr4,
		IncludePHP:  includePHP || hasPsr4,
		IncludeTwig: includeTwig,
	}, nil
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
