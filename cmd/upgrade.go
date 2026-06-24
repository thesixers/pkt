package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade pkt to the latest release",
	Long: `Upgrade pkt by downloading the latest pre-compiled binary from GitHub Releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔍 Checking for latest release on GitHub...")

		// Fetch latest release info
		resp, err := http.Get("https://api.github.com/repos/thesixers/pkt/releases/latest")
		if err != nil {
			return fmt.Errorf("failed to fetch latest release: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == 404 {
			return fmt.Errorf("no releases found yet for thesixers/pkt on GitHub. Please check back later")
		} else if resp.StatusCode != 200 {
			return fmt.Errorf("GitHub API returned status code: %d", resp.StatusCode)
		}

		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("failed to parse GitHub API response: %w", err)
		}

		fmt.Printf("📦 Found latest version: %s\n", release.TagName)

		// Determine the appropriate asset for this OS and Arch
		var downloadURL string
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			osName := runtime.GOOS
			archName := runtime.GOARCH

			// Handle common naming conventions
			if archName == "amd64" {
				archName = "x86_64"
			}

			// We look for both the exact OS and some variant of the Architecture
			if strings.Contains(name, osName) && (strings.Contains(name, runtime.GOARCH) || strings.Contains(name, archName)) {
				// Don't download tar.gz or zip source code if we can help it, unless it's the binary inside
				// For a direct binary replacement, we expect the raw binary, or a zip we'd have to extract.
				// Assuming standard distribution of raw binaries for now.
				if !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".zip") {
					downloadURL = asset.BrowserDownloadURL
					break
				}
			}
		}

		if downloadURL == "" {
			return fmt.Errorf("could not find a suitable binary for your system (%s/%s) in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
		}

		fmt.Printf("⬇️  Downloading %s...\n", downloadURL)

		// Download the binary
		binResp, err := http.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download binary: %w", err)
		}
		defer func() { _ = binResp.Body.Close() }()

		if binResp.StatusCode != 200 {
			return fmt.Errorf("download failed with status code: %d", binResp.StatusCode)
		}

		// Get current executable path
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine current executable path: %w", err)
		}

		// Create a temporary file alongside the current executable
		tmpFile := exePath + ".new"
		out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create temporary file (try running with sudo?): %w", err)
		}

		// Copy downloaded data to temp file
		if _, err := io.Copy(out, binResp.Body); err != nil {
			_ = out.Close()
			_ = os.Remove(tmpFile)
			return fmt.Errorf("failed to write binary: %w", err)
		}
		_ = out.Close()

		// Replace the old executable with the new one
		fmt.Println("🔄 Replacing old version...")
		if err := os.Rename(tmpFile, exePath); err != nil {
			// Fallback: try to remove original first (required on Windows, occasionally on Linux)
			_ = os.Remove(exePath)
			if renameErr := os.Rename(tmpFile, exePath); renameErr != nil {
				return fmt.Errorf("failed to replace executable. Try running with sudo. Error: %w", err)
			}
		}

		fmt.Printf("\n✨ Successfully upgraded pkt to %s!\n", release.TagName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
