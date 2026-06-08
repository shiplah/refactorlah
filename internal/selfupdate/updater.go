package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"refactorlah/internal/buildinfo"
)

type Updater struct {
	BuildInfo  buildinfo.Info
	Executable string
	Locator    ReleaseLocator
	Downloader AssetDownloader
	Stdout     io.Writer
	Stderr     io.Writer
}

type AssetDownloader interface {
	Download(ctx context.Context, assetURL string) ([]byte, error)
}

type CheckOptions struct {
	TargetVersion string
}

type CheckResult struct {
	CurrentVersion      string `json:"current_version"`
	CurrentDistribution string `json:"current_distribution"`
	ExecutablePath      string `json:"executable_path"`
	TargetVersion       string `json:"target_version"`
	ReleaseURL          string `json:"release_url"`
	AssetName           string `json:"asset_name"`
	UpdateAvailable     bool   `json:"update_available"`
	UpToDate            bool   `json:"up_to_date"`
	Downgrade           bool   `json:"downgrade"`
}

type ApplyResult struct {
	CheckResult
	Updated         bool `json:"updated"`
	Staged          bool `json:"staged"`
	RestartRequired bool `json:"restart_required"`
}

func NewUpdater() (*Updater, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("determine executable path: %w", err)
	}
	if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
		executablePath = resolvedPath
	}

	client := NewGitHubClient()
	return &Updater{
		BuildInfo:  buildinfo.Current(),
		Executable: executablePath,
		Locator:    client,
		Downloader: client,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}, nil
}

func (u *Updater) Check(ctx context.Context, options CheckOptions) (CheckResult, error) {
	release, err := u.lookupRelease(ctx, options.TargetVersion)
	if err != nil {
		return CheckResult{}, err
	}

	archiveName, err := releaseArchiveName(u.BuildInfo.GOOS, u.BuildInfo.GOARCH)
	if err != nil {
		return CheckResult{}, err
	}
	if _, ok := findAsset(release.Assets, archiveName); !ok {
		return CheckResult{}, fmt.Errorf("release %s does not contain %s", release.TagName, archiveName)
	}
	if _, ok := findAsset(release.Assets, checksumAssetName); !ok {
		return CheckResult{}, fmt.Errorf("release %s does not contain %s", release.TagName, checksumAssetName)
	}

	result := CheckResult{
		CurrentVersion:      u.BuildInfo.Version,
		CurrentDistribution: u.BuildInfo.Distribution,
		ExecutablePath:      u.Executable,
		TargetVersion:       release.TagName,
		ReleaseURL:          release.HTMLURL,
		AssetName:           archiveName,
	}

	if options.TargetVersion != "" {
		result.UpdateAvailable = u.BuildInfo.Version != release.TagName
		result.UpToDate = !result.UpdateAvailable
		if compare, ok := compareSemanticVersions(release.TagName, u.BuildInfo.Version); ok && compare < 0 {
			result.Downgrade = true
		}
		return result, nil
	}

	if u.BuildInfo.Version == release.TagName {
		result.UpToDate = true
		return result, nil
	}

	if compare, ok := compareSemanticVersions(release.TagName, u.BuildInfo.Version); ok {
		if compare > 0 {
			result.UpdateAvailable = true
			return result, nil
		}
		result.UpToDate = true
		result.Downgrade = compare < 0
		return result, nil
	}

	result.UpdateAvailable = true
	return result, nil
}

func (u *Updater) Apply(ctx context.Context, options CheckOptions) (ApplyResult, error) {
	if u.Downloader == nil {
		client, ok := u.Locator.(*GitHubClient)
		if !ok {
			return ApplyResult{}, errors.New("missing release downloader")
		}
		u.Downloader = client
	}

	check, err := u.Check(ctx, options)
	if err != nil {
		return ApplyResult{}, err
	}
	if !check.UpdateAvailable {
		return ApplyResult{CheckResult: check}, nil
	}

	release, err := u.lookupRelease(ctx, options.TargetVersion)
	if err != nil {
		return ApplyResult{}, err
	}

	archiveAsset, ok := findAsset(release.Assets, check.AssetName)
	if !ok {
		return ApplyResult{}, fmt.Errorf("release %s is missing %s", release.TagName, check.AssetName)
	}
	checksumAsset, ok := findAsset(release.Assets, checksumAssetName)
	if !ok {
		return ApplyResult{}, fmt.Errorf("release %s is missing %s", release.TagName, checksumAssetName)
	}

	tempDir, err := os.MkdirTemp("", "refactorlah-update-*")
	if err != nil {
		return ApplyResult{}, fmt.Errorf("create update workspace: %w", err)
	}

	cleanupTempDir := true
	defer func() {
		if cleanupTempDir {
			_ = os.RemoveAll(tempDir)
		}
	}()

	archiveContent, err := u.Downloader.Download(ctx, archiveAsset.BrowserDownloadURL)
	if err != nil {
		return ApplyResult{}, err
	}
	checksumContent, err := u.Downloader.Download(ctx, checksumAsset.BrowserDownloadURL)
	if err != nil {
		return ApplyResult{}, err
	}

	expectedDigest, err := expectedChecksum(checksumContent, archiveAsset.Name)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := verifyChecksum(archiveContent, expectedDigest); err != nil {
		return ApplyResult{}, err
	}

	extractedPath := filepath.Join(tempDir, binaryName(u.BuildInfo.GOOS))
	if err := extractBinaryFromArchive(archiveAsset.Name, archiveContent, extractedPath, binaryName(u.BuildInfo.GOOS)); err != nil {
		return ApplyResult{}, err
	}

	result := ApplyResult{CheckResult: check}
	if runtime.GOOS == "windows" {
		if err := u.launchWindowsReplacement(tempDir, extractedPath); err != nil {
			return ApplyResult{}, err
		}
		cleanupTempDir = false
		result.Updated = true
		result.Staged = true
		result.RestartRequired = true
		return result, nil
	}

	if err := u.replaceInPlace(extractedPath); err != nil {
		return ApplyResult{}, err
	}
	result.Updated = true
	return result, nil
}

func (u *Updater) lookupRelease(ctx context.Context, targetVersion string) (Release, error) {
	if u.Locator == nil {
		return Release{}, errors.New("missing release locator")
	}

	if targetVersion != "" {
		return u.Locator.ByTag(ctx, targetVersion)
	}

	return u.Locator.Latest(ctx)
}

func (u *Updater) replaceInPlace(sourcePath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open extracted binary: %w", err)
	}
	defer source.Close()

	mode := os.FileMode(0o755)
	if info, err := os.Stat(u.Executable); err == nil && info.Mode().Perm() != 0 {
		mode = info.Mode().Perm()
	}

	stagePath := filepath.Join(filepath.Dir(u.Executable), "."+filepath.Base(u.Executable)+".new")
	_ = os.Remove(stagePath)
	if err := writeExecutableFile(stagePath, source, mode); err != nil {
		return err
	}
	if err := os.Rename(stagePath, u.Executable); err != nil {
		_ = os.Remove(stagePath)
		return fmt.Errorf("replace executable: %w", err)
	}

	return nil
}

func (u *Updater) launchWindowsReplacement(tempDir string, sourcePath string) error {
	currentExecutable, err := os.Open(u.Executable)
	if err != nil {
		return fmt.Errorf("open current executable: %w", err)
	}
	defer currentExecutable.Close()

	helperPath := filepath.Join(tempDir, "refactorlah-update-helper.exe")
	if err := writeExecutableFile(helperPath, currentExecutable, 0o755); err != nil {
		return err
	}

	command := exec.Command(helperPath, ReplaceHelperCommand, "--target", u.Executable, "--source", sourcePath, "--cleanup-dir", tempDir)
	command.Stdout = u.Stdout
	command.Stderr = u.Stderr
	if err := command.Start(); err != nil {
		return fmt.Errorf("launch replacement helper: %w", err)
	}

	return nil
}

func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}

	return Asset{}, false
}
