package adapters

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type lookPathFunc func(string) (string, error)
type runCommandFunc func(context.Context, string, ...string) error

type Discovery struct {
	lookPath lookPathFunc
	run      runCommandFunc
}

func NewDiscovery() *Discovery {
	return &Discovery{
		lookPath: defaultLookPath,
		run:      defaultRun,
	}
}

func (d *Discovery) FindPHPAdapter(projectRoot string) (string, bool) {
	return d.FindAdapter("php", projectRoot)
}

func (d *Discovery) RequirePHPAdapter(ctx context.Context, projectRoot string) (string, error) {
	path, available := d.FindPHPAdapter(projectRoot)
	if !available {
		return "", fmt.Errorf("%w: native PHP/Twig support is unavailable in this build and no external refactorlah-php adapter was found", ErrAdapterFailure)
	}
	if err := d.preflight().Check(ctx, path); err != nil {
		return "", err
	}
	return path, nil
}

func (d *Discovery) FindPythonAdapter(projectRoot string) (string, bool) {
	return d.FindAdapter("python", projectRoot)
}

func (d *Discovery) RequirePythonAdapter(ctx context.Context, projectRoot string) (string, error) {
	path, available := d.FindPythonAdapter(projectRoot)
	if !available {
		return "", fmt.Errorf("%w: native Python support is unavailable in this build and no external refactorlah-python adapter was found", ErrAdapterFailure)
	}
	if err := d.preflight().Check(ctx, path); err != nil {
		return "", err
	}
	return path, nil
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

func (d *Discovery) preflight() Preflight {
	return Preflight{
		lookPath: d.lookPath,
		run:      d.run,
	}
}

func defaultLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func defaultRun(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
