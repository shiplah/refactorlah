package validation

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func pythonStaticChecks(projectRoot string) []validationCheck {
	checks := []validationCheck{}

	if ruffConfigured(projectRoot) {
		if tool, ok := projectTool(projectRoot, "ruff"); ok {
			checks = append(checks, validationCheck{name: "ruff", args: []string{tool, "check", "."}})
		}
	}

	if mypyConfigured(projectRoot) {
		if tool, ok := projectTool(projectRoot, "mypy"); ok {
			checks = append(checks, validationCheck{name: "mypy", args: []string{tool, "."}})
		}
	}

	return checks
}

func pythonTestChecks(projectRoot string) []validationCheck {
	if !pytestConfigured(projectRoot) {
		return nil
	}
	if tool, ok := projectTool(projectRoot, "pytest"); ok {
		return []validationCheck{{name: "pytest", args: []string{tool}}}
	}
	return nil
}

func ruffConfigured(projectRoot string) bool {
	if hasConfigFile(projectRoot, "ruff.toml", ".ruff.toml") {
		return true
	}
	return configContains(projectRoot, "pyproject.toml", "[tool.ruff")
}

func mypyConfigured(projectRoot string) bool {
	if hasConfigFile(projectRoot, "mypy.ini", ".mypy.ini") {
		return true
	}
	return configContains(projectRoot, "pyproject.toml", "[tool.mypy]") ||
		configContains(projectRoot, "setup.cfg", "[mypy]")
}

func pytestConfigured(projectRoot string) bool {
	if hasConfigFile(projectRoot, "pytest.ini") {
		return true
	}
	return configContains(projectRoot, "pyproject.toml", "[tool.pytest.ini_options]") ||
		configContains(projectRoot, "setup.cfg", "[tool:pytest]", "[pytest]") ||
		configContains(projectRoot, "tox.ini", "[pytest]")
}

func hasConfigFile(projectRoot string, names ...string) bool {
	for _, name := range names {
		info, err := os.Stat(filepath.Join(projectRoot, name))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func configContains(projectRoot string, name string, needles ...string) bool {
	content, err := os.ReadFile(filepath.Join(projectRoot, name))
	if err != nil {
		return false
	}
	text := string(content)
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func projectTool(projectRoot string, name string) (string, bool) {
	for _, candidate := range projectToolCandidates(projectRoot, name) {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, true
		}
	}

	path, err := exec.LookPath(name)
	return path, err == nil
}

func projectToolCandidates(projectRoot string, name string) []string {
	candidates := []string{
		filepath.Join(projectRoot, ".venv", "bin", name),
		filepath.Join(projectRoot, "venv", "bin", name),
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(projectRoot, ".venv", "Scripts", name+".exe"),
			filepath.Join(projectRoot, "venv", "Scripts", name+".exe"),
		)
	}
	return candidates
}
