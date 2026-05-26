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
		lookPath: exec.LookPath,
		run: func(ctx context.Context, name string, args ...string) error {
			command := exec.CommandContext(ctx, name, args...)
			output, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%w: %s", err, string(output))
			}
			return nil
		},
	}
}

func (d *Discovery) FindPHPAdapter(projectRoot string) (string, bool) {
	return d.FindAdapter("php", projectRoot)
}

func (d *Discovery) RequirePHPAdapter(ctx context.Context, projectRoot string) (string, error) {
	path, available := d.FindPHPAdapter(projectRoot)
	if !available {
		return "", fmt.Errorf("%w: PHP/Twig adapter is relevant but unavailable; build or install refactorlah-php", ErrAdapterFailure)
	}
	if err := d.checkPHPRuntime(ctx, path); err != nil {
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
		return "", fmt.Errorf("%w: Python adapter is relevant but unavailable; build or install refactorlah-python", ErrAdapterFailure)
	}
	if err := d.checkPythonRuntime(ctx, path); err != nil {
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

func (d *Discovery) checkPHPRuntime(ctx context.Context, adapterPath string) error {
	phpPath, err := d.runtimeLookPath("php")
	if err != nil {
		return fmt.Errorf("%w: PHP/Twig adapter requires php >= 8.2, but php was not found in PATH", ErrAdapterFailure)
	}
	if err := d.runtimeRun(ctx, phpPath, "-r", "exit(version_compare(PHP_VERSION, '8.2.0', '>=') ? 0 : 1);"); err != nil {
		return fmt.Errorf("%w: PHP/Twig adapter requires php >= 8.2", ErrAdapterFailure)
	}

	autoloadPath := filepath.Join(adapterRoot(adapterPath), "vendor", "autoload.php")
	if !fileExists(autoloadPath) {
		return fmt.Errorf("%w: PHP/Twig adapter dependencies are missing at %s; run composer install or rebuild refactorlah", ErrAdapterFailure, autoloadPath)
	}

	return nil
}

func (d *Discovery) checkPythonRuntime(ctx context.Context, adapterPath string) error {
	pythonPath, err := d.runtimeLookPath("python3")
	if err != nil {
		return fmt.Errorf("%w: Python adapter requires python3 >= 3.11, but python3 was not found in PATH", ErrAdapterFailure)
	}
	if err := d.runtimeRun(ctx, pythonPath, "-c", "import sys; raise SystemExit(0 if sys.version_info >= (3, 11) else 1)"); err != nil {
		return fmt.Errorf("%w: Python adapter requires python3 >= 3.11", ErrAdapterFailure)
	}

	entrypoint := filepath.Join(adapterRoot(adapterPath), "src", "analyze_command.py")
	if !fileExists(entrypoint) {
		return fmt.Errorf("%w: Python adapter files are missing at %s; rebuild refactorlah", ErrAdapterFailure, entrypoint)
	}

	return nil
}

func adapterRoot(adapterPath string) string {
	binDir := filepath.Dir(adapterPath)
	if filepath.Base(binDir) == "bin" {
		return filepath.Dir(binDir)
	}
	return binDir
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
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

func (d *Discovery) runtimeLookPath(name string) (string, error) {
	if d.lookPath == nil {
		return exec.LookPath(name)
	}
	return d.lookPath(name)
}

func (d *Discovery) runtimeRun(ctx context.Context, name string, args ...string) error {
	if d.run == nil {
		command := exec.CommandContext(ctx, name, args...)
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%w: %s", err, string(output))
		}
		return nil
	}
	return d.run(ctx, name, args...)
}
