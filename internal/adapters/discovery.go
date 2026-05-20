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

	if bundledPath, ok := bundledPHPAdapterPath(); ok {
		return bundledPath, true
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

func bundledPHPAdapterPath() (string, bool) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", false
	}

	executableDir := filepath.Dir(executablePath)
	candidates := []string{
		filepath.Join(executableDir, "libexec", "refactorlah-php", "bin", "refactorlah-php"),
		filepath.Join(executableDir, "refactorlah-php", "bin", "refactorlah-php"),
	}

	for _, candidate := range candidates {
		if isExecutable(candidate) {
			return candidate, true
		}
	}

	return "", false
}
