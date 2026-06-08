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

	"github.com/NickSdot/refactorlah/internal/buildinfo"
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
	SelfUpdateSupported bool   `json:"self_update_supported"`
	UpdateInstructions  string `json:"update_instructions,omitempty"`
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

type UpdatePlan struct {
	CheckResult   CheckResult
	archiveAsset  Asset
	checksumAsset Asset
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
	plan, err := u.Plan(ctx, options)
	if err != nil {
		return CheckResult{}, err
	}

	return plan.CheckResult, nil
}

func (u *Updater) Plan(ctx context.Context, options CheckOptions) (UpdatePlan, error) {
	release, err := u.lookupRelease(ctx, options.TargetVersion)
	if err != nil {
		return UpdatePlan{}, err
	}

	result := CheckResult{
		CurrentVersion:      u.BuildInfo.Version,
		CurrentDistribution: u.BuildInfo.Distribution,
		ExecutablePath:      u.Executable,
		TargetVersion:       release.TagName,
		ReleaseURL:          release.HTMLURL,
		SelfUpdateSupported: selfUpdateSupported(u.BuildInfo.Distribution),
		UpdateInstructions:  updateInstructions(u.BuildInfo.Distribution),
	}

	result = classifyVersionState(result, options.TargetVersion != "")
	if !result.SelfUpdateSupported {
		return UpdatePlan{CheckResult: result}, nil
	}

	archiveName, err := releaseArchiveName(u.BuildInfo.GOOS, u.BuildInfo.GOARCH)
	if err != nil {
		return UpdatePlan{}, err
	}
	archiveAsset, ok := findAsset(release.Assets, archiveName)
	if !ok {
		return UpdatePlan{}, fmt.Errorf("release %s does not contain %s", release.TagName, archiveName)
	}
	checksumAsset, ok := findAsset(release.Assets, checksumAssetName)
	if !ok {
		return UpdatePlan{}, fmt.Errorf("release %s does not contain %s", release.TagName, checksumAssetName)
	}

	result.AssetName = archiveName
	plan := UpdatePlan{
		CheckResult:   result,
		archiveAsset:  archiveAsset,
		checksumAsset: checksumAsset,
	}

	return plan, nil
}

func classifyVersionState(result CheckResult, explicitTarget bool) CheckResult {
	if explicitTarget {
		result.UpdateAvailable = result.CurrentVersion != result.TargetVersion
		result.UpToDate = !result.UpdateAvailable
		if compare, ok := compareSemanticVersions(result.TargetVersion, result.CurrentVersion); ok {
			result.Downgrade = compare < 0
		}
		return result
	}

	compare, ok := compareSemanticVersions(result.TargetVersion, result.CurrentVersion)
	if !ok {
		result.UpdateAvailable = result.CurrentVersion != result.TargetVersion
		result.UpToDate = !result.UpdateAvailable
		return result
	}

	switch {
	case compare > 0:
		result.UpdateAvailable = true
	case compare == 0:
		result.UpToDate = true
	default:
		result.UpToDate = true
		result.Downgrade = true
	}

	return result
}

func (u *Updater) Apply(ctx context.Context, options CheckOptions) (ApplyResult, error) {
	plan, err := u.Plan(ctx, options)
	if err != nil {
		return ApplyResult{}, err
	}

	return u.ApplyPlan(ctx, plan)
}

func (u *Updater) ApplyPlan(ctx context.Context, plan UpdatePlan) (ApplyResult, error) {
	if !plan.CheckResult.UpdateAvailable {
		return ApplyResult{CheckResult: plan.CheckResult}, nil
	}
	if !plan.CheckResult.SelfUpdateSupported {
		return ApplyResult{CheckResult: plan.CheckResult}, nil
	}

	downloader, err := u.releaseDownloader()
	if err != nil {
		return ApplyResult{}, err
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

	archiveContent, err := downloader.Download(ctx, plan.archiveAsset.BrowserDownloadURL)
	if err != nil {
		return ApplyResult{}, err
	}
	checksumContent, err := downloader.Download(ctx, plan.checksumAsset.BrowserDownloadURL)
	if err != nil {
		return ApplyResult{}, err
	}

	expectedDigest, err := expectedChecksum(checksumContent, plan.archiveAsset.Name)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := verifyChecksum(archiveContent, expectedDigest); err != nil {
		return ApplyResult{}, err
	}

	extractedPath := filepath.Join(tempDir, binaryName(u.BuildInfo.GOOS))
	if err := extractBinaryFromArchive(plan.archiveAsset.Name, archiveContent, extractedPath, binaryName(u.BuildInfo.GOOS)); err != nil {
		return ApplyResult{}, err
	}

	result := ApplyResult{CheckResult: plan.CheckResult}
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

func (u *Updater) releaseDownloader() (AssetDownloader, error) {
	if u.Downloader != nil {
		return u.Downloader, nil
	}

	client, ok := u.Locator.(*GitHubClient)
	if !ok {
		return nil, errors.New("missing release downloader")
	}

	return client, nil
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

func selfUpdateSupported(distribution string) bool {
	return distribution == buildinfo.DistributionGitHubRelease
}

func updateInstructions(distribution string) string {
	switch distribution {
	case buildinfo.DistributionGitHubRelease:
		return ""
	case buildinfo.DistributionGoInstall:
		return "Update Go installs by rerunning: go install github.com/NickSdot/refactorlah/cmd/refactorlah@latest"
	case buildinfo.DistributionSourceInstall:
		return "Update source checkouts by pulling the latest changes and rerunning bin/install.sh."
	default:
		return "Install a GitHub release binary to use refactorlah update, or rebuild this source install manually."
	}
}
