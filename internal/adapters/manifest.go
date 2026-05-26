package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Manifest struct {
	Name       string          `json:"name"`
	Executable string          `json:"executable"`
	Runtime    RuntimeManifest `json:"runtime"`
}

type RuntimeManifest struct {
	Command        string `json:"command"`
	MinimumVersion string `json:"minimumVersion"`
}

func LoadManifest(adapterRoot string) (Manifest, error) {
	manifestPath := filepath.Join(adapterRoot, "adapter.json")
	contents, err := os.ReadFile(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("%w: adapter manifest is missing at %s", ErrAdapterFailure, manifestPath)
	}

	var manifest Manifest
	if err := json.Unmarshal(contents, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("%w: adapter manifest is invalid at %s: %v", ErrAdapterFailure, manifestPath, err)
	}
	if manifest.Name == "" || manifest.Executable == "" || manifest.Runtime.Command == "" || manifest.Runtime.MinimumVersion == "" {
		return Manifest{}, fmt.Errorf("%w: adapter manifest is incomplete at %s", ErrAdapterFailure, manifestPath)
	}

	return manifest, nil
}

type Preflight struct {
	lookPath lookPathFunc
	run      runCommandFunc
}

func (p Preflight) Check(ctx context.Context, adapterPath string) error {
	root := adapterRoot(adapterPath)
	manifest, err := LoadManifest(root)
	if err != nil {
		return err
	}

	return p.checkRuntime(ctx, manifest)
}

func (p Preflight) runtimeLookPath(name string) (string, error) {
	if p.lookPath == nil {
		return defaultLookPath(name)
	}
	return p.lookPath(name)
}

func (p Preflight) runtimeRun(ctx context.Context, name string, args ...string) error {
	if p.run == nil {
		return defaultRun(ctx, name, args...)
	}
	return p.run(ctx, name, args...)
}

func (p Preflight) checkRuntime(ctx context.Context, manifest Manifest) error {
	runtimePath, err := p.runtimeLookPath(manifest.Runtime.Command)
	if err != nil {
		return fmt.Errorf("%w: %s adapter requires %s >= %s, but %s was not found in PATH", ErrAdapterFailure, manifest.Name, manifest.Runtime.Command, manifest.Runtime.MinimumVersion, manifest.Runtime.Command)
	}

	checkArgs, err := runtimeVersionCheck(manifest.Runtime)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAdapterFailure, err)
	}
	if err := p.runtimeRun(ctx, runtimePath, checkArgs...); err != nil {
		return fmt.Errorf("%w: %s adapter requires %s >= %s", ErrAdapterFailure, manifest.Name, manifest.Runtime.Command, manifest.Runtime.MinimumVersion)
	}

	return nil
}

func runtimeVersionCheck(runtime RuntimeManifest) ([]string, error) {
	switch runtime.Command {
	case "php":
		return []string{
			"-r",
			fmt.Sprintf("exit(version_compare(PHP_VERSION, '%s', '>=') ? 0 : 1);", runtime.MinimumVersion),
		}, nil
	case "python3":
		return []string{
			"-c",
			fmt.Sprintf("import sys; raise SystemExit(0 if sys.version_info >= tuple(map(int, %q.split('.'))) else 1)", runtime.MinimumVersion),
		}, nil
	default:
		return nil, fmt.Errorf("no runtime version check is implemented for %q", runtime.Command)
	}
}

func adapterRoot(adapterPath string) string {
	binDir := filepath.Dir(adapterPath)
	if filepath.Base(binDir) == "bin" {
		return filepath.Dir(binDir)
	}
	return binDir
}
