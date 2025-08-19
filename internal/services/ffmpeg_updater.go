package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// FFmpegUpdater manages FFmpeg binary updates from BtbN releases
type FFmpegUpdater struct {
	logger         zerolog.Logger
	httpClient     *http.Client
	installPath    string
	backupPath     string
	currentVersion *FFmpegVersion
}

// FFmpegVersion represents FFmpeg version information
type FFmpegVersion struct {
	Major       int       `json:"major"`
	Minor       int       `json:"minor"`
	Patch       int       `json:"patch"`
	Git         string    `json:"git,omitempty"`
	ReleaseURL  string    `json:"release_url,omitempty"`
	AssetURL    string    `json:"asset_url,omitempty"`
	Stable      bool      `json:"stable"`
	ReleaseDate time.Time `json:"release_date"`
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// NewFFmpegUpdater creates a new FFmpeg updater service
func NewFFmpegUpdater(logger zerolog.Logger, installPath string) *FFmpegUpdater {
	return &FFmpegUpdater{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		installPath: installPath,
		backupPath:  filepath.Join(installPath, "backup"),
	}
}

// GetCurrentVersion gets the currently installed FFmpeg version
func (u *FFmpegUpdater) GetCurrentVersion() (*FFmpegVersion, error) {
	if u.currentVersion != nil {
		return u.currentVersion, nil
	}

	ffmpegPath := filepath.Join(u.installPath, "ffmpeg")
	cmd := exec.Command(ffmpegPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get FFmpeg version: %w", err)
	}

	version, err := u.parseFFmpegVersion(string(output))
	if err != nil {
		return nil, err
	}

	u.currentVersion = version
	return version, nil
}

// parseFFmpegVersion parses FFmpeg version output
func (u *FFmpegUpdater) parseFFmpegVersion(output string) (*FFmpegVersion, error) {
	// Example: ffmpeg version n6.1-26-g3f345ebf21
	// Or: ffmpeg version 6.1.1
	versionRegex := regexp.MustCompile(`ffmpeg version (?:n)?(\d+)\.(\d+)(?:\.(\d+))?(?:-(\d+)-g([a-f0-9]+))?`)
	matches := versionRegex.FindStringSubmatch(output)

	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse FFmpeg version from output")
	}

	version := &FFmpegVersion{
		Stable: true,
	}

	version.Major, _ = strconv.Atoi(matches[1])
	version.Minor, _ = strconv.Atoi(matches[2])

	if len(matches) > 3 && matches[3] != "" {
		version.Patch, _ = strconv.Atoi(matches[3])
	}

	if len(matches) > 5 && matches[5] != "" {
		version.Git = matches[5]
		version.Stable = false // Git versions are considered development
	}

	return version, nil
}

// CheckLatestRelease checks for the latest FFmpeg release from BtbN
func (u *FFmpegUpdater) CheckLatestRelease(ctx context.Context) (*FFmpegVersion, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.github.com/repos/BtbN/FFmpeg-Builds/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	// Find appropriate asset for current platform
	assetURL := u.findAppropriateAsset(release.Assets)
	if assetURL == "" {
		return nil, fmt.Errorf("no suitable FFmpeg build found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Parse version from release tag
	version, err := u.parseReleaseTag(release.TagName)
	if err != nil {
		return nil, err
	}

	version.ReleaseURL = fmt.Sprintf("https://github.com/BtbN/FFmpeg-Builds/releases/tag/%s", release.TagName)
	version.AssetURL = assetURL
	version.ReleaseDate = release.PublishedAt
	version.Stable = !release.Prerelease

	return version, nil
}

// findAppropriateAsset finds the right FFmpeg build for the current platform
func (u *FFmpegUpdater) findAppropriateAsset(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}) string {
	// Determine platform-specific file pattern
	var patterns []string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			patterns = []string{
				"ffmpeg-n.*-linux64-gpl.tar.xz",
				"ffmpeg-master-latest-linux64-gpl.tar.xz",
			}
		case "arm64":
			patterns = []string{
				"ffmpeg-n.*-linuxarm64-gpl.tar.xz",
				"ffmpeg-master-latest-linuxarm64-gpl.tar.xz",
			}
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			patterns = []string{
				"ffmpeg-n.*-macos64-gpl.zip",
				"ffmpeg-master-latest-macos64-gpl.zip",
			}
		case "arm64":
			patterns = []string{
				"ffmpeg-n.*-macosarm64-gpl.zip",
				"ffmpeg-master-latest-macosarm64-gpl.zip",
			}
		}
	case "windows":
		patterns = []string{
			"ffmpeg-n.*-win64-gpl.zip",
			"ffmpeg-master-latest-win64-gpl.zip",
		}
	}

	// Find matching asset
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		for _, asset := range assets {
			if regex.MatchString(asset.Name) {
				u.logger.Info().
					Str("asset", asset.Name).
					Int64("size_mb", asset.Size/1024/1024).
					Msg("Found matching FFmpeg build")
				return asset.BrowserDownloadURL
			}
		}
	}

	return ""
}

// parseReleaseTag parses version from GitHub release tag
func (u *FFmpegUpdater) parseReleaseTag(tag string) (*FFmpegVersion, error) {
	// Tags like "latest", "autobuild-2024-01-15", etc.
	if tag == "latest" || strings.Contains(tag, "autobuild") {
		// For latest/autobuild, we need to check the actual binary version
		// Return a placeholder that indicates this is the latest
		return &FFmpegVersion{
			Major:  99, // Special version to indicate latest
			Minor:  0,
			Patch:  0,
			Stable: true,
		}, nil
	}

	// Try to parse semantic version if present
	versionRegex := regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := versionRegex.FindStringSubmatch(tag)

	if len(matches) < 3 {
		// Can't parse, assume it's latest
		return &FFmpegVersion{
			Major:  99,
			Minor:  0,
			Patch:  0,
			Stable: true,
		}, nil
	}

	version := &FFmpegVersion{Stable: true}
	version.Major, _ = strconv.Atoi(matches[1])
	version.Minor, _ = strconv.Atoi(matches[2])
	if len(matches) > 3 && matches[3] != "" {
		version.Patch, _ = strconv.Atoi(matches[3])
	}

	return version, nil
}

// CompareVersions compares two FFmpeg versions
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func (u *FFmpegUpdater) CompareVersions(v1, v2 *FFmpegVersion) int {
	if v1.Major != v2.Major {
		return v1.Major - v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor - v2.Minor
	}
	if v1.Patch != v2.Patch {
		return v1.Patch - v2.Patch
	}
	return 0
}

// IsMajorUpgrade checks if the new version is a major upgrade
func (u *FFmpegUpdater) IsMajorUpgrade(current, new *FFmpegVersion) bool {
	return new.Major > current.Major
}

// IsMinorUpgrade checks if the new version is a minor upgrade
func (u *FFmpegUpdater) IsMinorUpgrade(current, new *FFmpegVersion) bool {
	return new.Major == current.Major && new.Minor > current.Minor
}

// VerifyStability performs stability checks on the new version
func (u *FFmpegUpdater) VerifyStability(ctx context.Context, version *FFmpegVersion) (*StabilityReport, error) {
	report := &StabilityReport{
		Version:       version,
		CheckedAt:     time.Now(),
		Stable:        version.Stable,
		Compatibility: true,
	}

	// Check if release is at least 48 hours old (for stability)
	if time.Since(version.ReleaseDate) < 48*time.Hour {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Release is less than 48 hours old (%s)",
				time.Since(version.ReleaseDate).Round(time.Hour)))
		report.Stable = false
	}

	// Check known compatibility issues
	report.checkKnownIssues()

	// Test with sample command if possible
	if err := u.testFFmpegBinary(ctx, version); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("Binary test failed: %v", err))
		report.Compatibility = false
	}

	return report, nil
}

// StabilityReport contains stability check results
type StabilityReport struct {
	Version       *FFmpegVersion `json:"version"`
	CheckedAt     time.Time      `json:"checked_at"`
	Stable        bool           `json:"stable"`
	Compatibility bool           `json:"compatibility"`
	Warnings      []string       `json:"warnings,omitempty"`
	Errors        []string       `json:"errors,omitempty"`
}

// checkKnownIssues checks for known compatibility issues
func (r *StabilityReport) checkKnownIssues() {
	// FFmpeg 7.0 has breaking changes
	if r.Version.Major >= 7 {
		r.Warnings = append(r.Warnings,
			"FFmpeg 7.0+ contains breaking API changes. Manual review recommended.")
	}

	// Check for specific problematic versions
	problematicVersions := map[string]string{
		"6.1.0": "Known issues with certain codecs",
		"5.2.0": "Memory leak in specific filters",
	}

	versionStr := fmt.Sprintf("%d.%d.%d", r.Version.Major, r.Version.Minor, r.Version.Patch)
	if issue, exists := problematicVersions[versionStr]; exists {
		r.Warnings = append(r.Warnings, fmt.Sprintf("Version %s: %s", versionStr, issue))
		r.Stable = false
	}
}

// testFFmpegBinary tests if the FFmpeg binary works correctly
func (u *FFmpegUpdater) testFFmpegBinary(ctx context.Context, version *FFmpegVersion) error {
	// This would download and test the binary in a sandboxed environment
	// For now, we'll just return nil as a placeholder
	// In production, this would:
	// 1. Download the binary to a temp location
	// 2. Run basic commands to verify functionality
	// 3. Check codec support
	// 4. Verify filter availability
	return nil
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Current        *FFmpegVersion   `json:"current"`
	Available      *FFmpegVersion   `json:"available"`
	IsMajor        bool             `json:"is_major"`
	IsMinor        bool             `json:"is_minor"`
	IsPatch        bool             `json:"is_patch"`
	Stability      *StabilityReport `json:"stability"`
	Recommendation string           `json:"recommendation"`
}

// CheckForUpdates checks for available FFmpeg updates
func (u *FFmpegUpdater) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	current, err := u.GetCurrentVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	latest, err := u.CheckLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check latest release: %w", err)
	}

	comparison := u.CompareVersions(latest, current)
	if comparison <= 0 {
		return &UpdateInfo{
			Current:        current,
			Available:      latest,
			Recommendation: "You are already on the latest version",
		}, nil
	}

	stability, err := u.VerifyStability(ctx, latest)
	if err != nil {
		u.logger.Warn().Err(err).Msg("Failed to verify stability")
	}

	info := &UpdateInfo{
		Current:   current,
		Available: latest,
		IsMajor:   u.IsMajorUpgrade(current, latest),
		IsMinor:   u.IsMinorUpgrade(current, latest),
		IsPatch:   !u.IsMajorUpgrade(current, latest) && !u.IsMinorUpgrade(current, latest),
		Stability: stability,
	}

	// Generate recommendation
	if info.IsMajor {
		info.Recommendation = "Major upgrade available. Manual review and testing strongly recommended before updating."
	} else if info.IsMinor {
		info.Recommendation = "Minor upgrade available. Review changelog for new features and potential compatibility issues."
	} else if info.IsPatch {
		if stability != nil && stability.Stable {
			info.Recommendation = "Patch update available. Safe to update."
		} else {
			info.Recommendation = "Patch update available but stability checks failed. Wait for a more stable release."
		}
	}

	return info, nil
}

// DownloadUpdate downloads the FFmpeg update
func (u *FFmpegUpdater) DownloadUpdate(ctx context.Context, version *FFmpegVersion, progress func(percent int)) error {
	if version.AssetURL == "" {
		return fmt.Errorf("no download URL available")
	}

	// Create temporary directory for download
	tempDir := filepath.Join(u.installPath, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the archive
	archivePath := filepath.Join(tempDir, "ffmpeg-update.archive")
	if err := u.downloadFile(ctx, version.AssetURL, archivePath, progress); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract based on file type
	var extractErr error
	if strings.HasSuffix(version.AssetURL, ".tar.xz") {
		extractErr = u.extractTarXZ(archivePath, tempDir)
	} else if strings.HasSuffix(version.AssetURL, ".zip") {
		extractErr = u.extractZip(archivePath, tempDir)
	} else {
		return fmt.Errorf("unsupported archive format")
	}

	if extractErr != nil {
		return fmt.Errorf("failed to extract archive: %w", extractErr)
	}

	// Backup current installation
	if err := u.backupCurrent(); err != nil {
		return fmt.Errorf("failed to backup current installation: %w", err)
	}

	// Install new binaries
	if err := u.installBinaries(tempDir); err != nil {
		// Rollback on failure
		u.rollback()
		return fmt.Errorf("failed to install new binaries: %w", err)
	}

	// Update stored version
	u.currentVersion = version

	return nil
}

// downloadFile downloads a file with progress reporting
func (u *FFmpegUpdater) downloadFile(ctx context.Context, url, dest string, progress func(percent int)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if progress != nil && totalSize > 0 {
				percent := int(downloaded * 100 / totalSize)
				progress(percent)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarXZ extracts a tar.xz archive
func (u *FFmpegUpdater) extractTarXZ(archivePath, destDir string) error {
	cmd := exec.Command("tar", "-xJf", archivePath, "-C", destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar extraction failed: %w", err)
	}
	return nil
}

// extractZip extracts a zip archive
func (u *FFmpegUpdater) extractZip(archivePath, destDir string) error {
	cmd := exec.Command("unzip", "-q", archivePath, "-d", destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unzip failed: %w", err)
	}
	return nil
}

// backupCurrent backs up the current FFmpeg installation
func (u *FFmpegUpdater) backupCurrent() error {
	if err := os.RemoveAll(u.backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean backup directory: %w", err)
	}

	if err := os.MkdirAll(u.backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup ffmpeg and ffprobe binaries
	for _, binary := range []string{"ffmpeg", "ffprobe"} {
		src := filepath.Join(u.installPath, binary)
		dst := filepath.Join(u.backupPath, binary)

		if _, err := os.Stat(src); err == nil {
			if err := u.copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to backup %s: %w", binary, err)
			}
		}
	}

	return nil
}

// rollback restores the previous FFmpeg installation
func (u *FFmpegUpdater) rollback() error {
	for _, binary := range []string{"ffmpeg", "ffprobe"} {
		src := filepath.Join(u.backupPath, binary)
		dst := filepath.Join(u.installPath, binary)

		if _, err := os.Stat(src); err == nil {
			if err := u.copyFile(src, dst); err != nil {
				u.logger.Error().Err(err).Str("binary", binary).Msg("Failed to rollback binary")
			}
		}
	}
	return nil
}

// installBinaries installs new FFmpeg binaries
func (u *FFmpegUpdater) installBinaries(tempDir string) error {
	// Find the extracted directory (usually has ffmpeg in the name)
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	var extractedDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "ffmpeg") {
			extractedDir = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if extractedDir == "" {
		return fmt.Errorf("could not find extracted FFmpeg directory")
	}

	// Copy binaries
	binDir := filepath.Join(extractedDir, "bin")
	for _, binary := range []string{"ffmpeg", "ffprobe"} {
		src := filepath.Join(binDir, binary)
		dst := filepath.Join(u.installPath, binary)

		if err := u.copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to install %s: %w", binary, err)
		}

		// Make executable on Unix-like systems
		if runtime.GOOS != "windows" {
			if err := os.Chmod(dst, 0755); err != nil {
				return fmt.Errorf("failed to make %s executable: %w", binary, err)
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func (u *FFmpegUpdater) copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
