package adapters

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Discovery struct{}

func NewDiscovery() *Discovery {
	return &Discovery{}
}

func (d *Discovery) FindPHPAdapter(projectRoot string) (string, bool) {
	if envPath := os.Getenv("REFACTORLAH_PHP_ADAPTER"); envPath != "" {
		if isExecutable(envPath) {
			return envPath, true
		}
	}

	if path, err := exec.LookPath("refactorlah-php"); err == nil {
		return path, true
	}

	localPath := filepath.Join(projectRoot, "adapters", "php", "bin", "refactorlah-php")
	if isExecutable(localPath) {
		return localPath, true
	}

	return "", false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}
