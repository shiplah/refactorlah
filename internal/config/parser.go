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
		Checks  [][]string `json:"checks"`
		Tests   [][]string `json:"tests"`
	}
	if err := json.Unmarshal(content, &payload); err != nil {
		return Config{}, err
	}

	return Config{
		Include: nonEmptyStrings(payload.Include),
		Exclude: nonEmptyStrings(payload.Exclude),
		Checks:  nonEmptyCommands(payload.Checks),
		Tests:   nonEmptyCommands(payload.Tests),
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

func nonEmptyCommands(commands [][]string) [][]string {
	result := make([][]string, 0, len(commands))
	for _, command := range commands {
		cleaned := make([]string, 0, len(command))
		for _, argument := range command {
			argument = strings.TrimSpace(argument)
			if argument != "" {
				cleaned = append(cleaned, argument)
			}
		}
		if len(cleaned) > 0 {
			result = append(result, cleaned)
		}
	}
	return result
}
