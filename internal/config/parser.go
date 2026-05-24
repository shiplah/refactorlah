package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func readConfigFile(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var payload struct {
		Include []string `json:"include"`
		Exclude []string `json:"exclude"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return Config{}, err
	}

	return Config{
		Include: nonEmptyStrings(payload.Include),
		Exclude: nonEmptyStrings(payload.Exclude),
	}, nil
}

func nonEmptyStrings(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(filepath.ToSlash(item))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
