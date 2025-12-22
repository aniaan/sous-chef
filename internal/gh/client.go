package gh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Release struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	Assets      []Asset   `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"` // Custom field, optional
}

const githubAPIBaseURL = "https://api.github.com"

// Client is a simple GitHub API client
type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetReleaseByTag fetches a specific release by tag
func (c *Client) GetReleaseByTag(repo, tag string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/tags/%s", githubAPIBaseURL, repo, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "sous-chef")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// GetAssetChecksum fetches the checksum (digest) for a specific asset if available
// Returns empty string if not found or no digest present
func (c *Client) GetAssetChecksum(repo, tag, filename string) (string, error) {
	release, err := c.GetReleaseByTag(repo, tag)
	if err != nil {
		return "", err
	}

	for _, asset := range release.Assets {
		if asset.Name == filename {
			if asset.Digest != "" {
				return asset.Digest[7:], nil
			}
			return "", nil
		}
	}
	return "", nil // Asset not found or no digest
}

// ListReleases fetches the latest releases for a repository
func (c *Client) ListReleases(repo string) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases", githubAPIBaseURL, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "sous-chef")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// DownloadReleaseAsset downloads a release asset to a destination path
func (c *Client) DownloadReleaseAsset(repo, tag, filename, destPath string) error {
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, filename)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "sous-chef")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
