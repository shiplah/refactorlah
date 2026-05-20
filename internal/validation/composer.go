package validation

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

type composerConfig struct {
	Scripts map[string]any `json:"scripts"`
}

func composerAvailable() bool {
	_, err := exec.LookPath("composer")
	return err == nil
}

func composerHasTestScript(projectRoot string) bool {
	bytes, err := os.ReadFile(filepath.Join(projectRoot, "composer.json"))
	if err != nil {
		return false
	}

	var config composerConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return false
	}
	_, ok := config.Scripts["test"]
	return ok
}
