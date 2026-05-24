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
	return d.FindAdapter("php", projectRoot)
}

func (d *Discovery) FindPythonAdapter(projectRoot string) (string, bool) {
	return d.FindAdapter("python", projectRoot)
}

func (d *Discovery) FindAdapter(name string, projectRoot string) (string, bool) {
	executableName := "refactorlah-" + name
	if envPath := os.Getenv(adapterEnvName(name)); envPath != "" {
		if isExecutable(envPath) {
			return envPath, true
		}
	}

	if bundledPath, ok := bundledAdapterPath(name); ok {
		return bundledPath, true
	}

	if path, err := exec.LookPath(executableName); err == nil {
		return path, true
	}

	localPath := filepath.Join(projectRoot, "adapters", name, "bin", executableName)
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

func bundledAdapterPath(name string) (string, bool) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", false
	}
	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executablePath = resolvedPath
	}

	executableDir := filepath.Dir(executablePath)
	executableName := "refactorlah-" + name
	candidates := []string{
		filepath.Join(executableDir, "libexec", executableName, "bin", executableName),
		filepath.Join(executableDir, executableName, "bin", executableName),
	}

	for _, candidate := range candidates {
		if isExecutable(candidate) {
			return candidate, true
		}
	}

	return "", false
}

func adapterEnvName(name string) string {
	switch name {
	case "php":
		return "REFACTORLAH_PHP_ADAPTER"
	case "python":
		return "REFACTORLAH_PYTHON_ADAPTER"
	default:
		return ""
	}
}
