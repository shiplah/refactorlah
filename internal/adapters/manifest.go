package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Manifest struct {
	Name          string          `json:"name"`
	Executable    string          `json:"executable"`
	Runtime       RuntimeManifest `json:"runtime"`
	RequiredFiles []string        `json:"requiredFiles"`
}

type RuntimeManifest struct {
	Command        string   `json:"command"`
	MinimumVersion string   `json:"minimumVersion"`
	VersionCheck   []string `json:"versionCheck"`
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

	runtimePath, err := p.runtimeLookPath(manifest.Runtime.Command)
	if err != nil {
		return fmt.Errorf("%w: %s adapter requires %s >= %s, but %s was not found in PATH", ErrAdapterFailure, manifest.Name, manifest.Runtime.Command, manifest.Runtime.MinimumVersion, manifest.Runtime.Command)
	}
	if len(manifest.Runtime.VersionCheck) > 0 {
		if err := p.runtimeRun(ctx, runtimePath, manifest.Runtime.VersionCheck...); err != nil {
			return fmt.Errorf("%w: %s adapter requires %s >= %s", ErrAdapterFailure, manifest.Name, manifest.Runtime.Command, manifest.Runtime.MinimumVersion)
		}
	}

	for _, requiredFile := range manifest.RequiredFiles {
		path := filepath.Join(root, filepath.FromSlash(requiredFile))
		if !fileExists(path) {
			return fmt.Errorf("%w: %s adapter dependency is missing at %s; install adapter dependencies or rebuild refactorlah", ErrAdapterFailure, manifest.Name, path)
		}
	}

	return nil
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
