package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aniaan/sous-chef/internal/gh"
	"github.com/aniaan/sous-chef/internal/registry"
	"github.com/aniaan/sous-chef/internal/util"
)

type Context struct {
	Version  string
	Platform string
	Arch     string
}

// Install handles the download and installation of a tool
func Install(plugin *registry.PluginConfig, version, installDir string) error {
	plat, arch, err := util.GetSystemInfo()
	if err != nil {
		return err
	}

	// Map platform/arch to plugin specific strings
	platStr := string(plat)
	if val, ok := plugin.PlatformMap[plat]; ok {
		platStr = val
	}
	archStr := string(arch)
	if val, ok := plugin.ArchMap[arch]; ok {
		archStr = val
	}

	ctx := Context{
		Version:  version,
		Platform: platStr,
		Arch:     archStr,
	}

	// Render filename
	filename, err := renderTemplate(plugin.AssetTemplate, ctx)
	if err != nil {
		return fmt.Errorf("failed to render filename: %w", err)
	}

	// Determine GitHub Tag
	tag := version
	if plugin.RecoverVersion != nil {
		tag = plugin.RecoverVersion(version)
	}

	// Download to temp file
	tempDir, err := os.MkdirTemp("", "sous-chef")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // Clean up

	downloadPath := filepath.Join(tempDir, filename)
	fmt.Printf("Downloading %s/%s@%s...\n", plugin.Repo, filename, tag)

	client := gh.NewClient()
	if err := client.DownloadReleaseAsset(plugin.Repo, tag, filename, downloadPath); err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	// Verify Checksum
	checksum, err := client.GetAssetChecksum(plugin.Repo, tag, filename)
	if err != nil {
		fmt.Printf("Warning: failed to get checksum: %v\n", err)
	}

	if checksum != "" {
		fmt.Printf("Verifying checksum for %s...\n", filename)
		if err := verifyChecksum(downloadPath, checksum); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		fmt.Println("Checksum verified.")
	} else {
		fmt.Println("No checksum found in GitHub API, skipping verification.")
	}

	// Resolve relative binary path early
	relBinPath, err := renderTemplate(plugin.RelativeBinPathTemplate, ctx)
	if err != nil {
		return fmt.Errorf("failed to render relative bin path: %w", err)
	}

	// Extract
	fmt.Printf("Extracting to %s...\n", installDir)
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		if err := util.ExtractTarGz(downloadPath, installDir, plugin.StripComponents); err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".gz") {
		if err := util.ExtractGz(downloadPath, filepath.Join(installDir, relBinPath)); err != nil {
			return err
		}
	} else if strings.HasSuffix(filename, ".zip") {
		if err := util.ExtractZip(downloadPath, installDir, plugin.StripComponents); err != nil {
			return err
		}
	} else {
		// Assume it's a raw executable (like shfmt/gofumpt)
		targetPath := filepath.Join(installDir, relBinPath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := util.CopyFile(downloadPath, targetPath); err != nil {
			return err
		}
	}

	// Locate binary and move to bin/
	// (relBinPath is already rendered above)

	srcBin := filepath.Join(installDir, relBinPath)
	destBinDir := filepath.Join(installDir, "bin")
	if err := os.MkdirAll(destBinDir, 0o755); err != nil {
		return err
	}
	destBin := filepath.Join(destBinDir, plugin.Cmd)

	// Check if source exists
	if _, err := os.Stat(srcBin); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s", srcBin)
	}

	if srcBin != destBin {
		fmt.Printf("Moving %s to %s...\n", srcBin, destBin)
		if err := os.Rename(srcBin, destBin); err != nil {
			return err
		}
	}

	// Chmod +x
	if err := os.Chmod(destBin, 0o755); err != nil {
		return err
	}

	return nil
}

func renderTemplate(tmplStr string, data any) (string, error) {
	tmpl, err := template.New("filename").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func verifyChecksum(filepath, expected string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expected {
		return fmt.Errorf("expected %s, got %s", expected, actual)
	}
	return nil
}
