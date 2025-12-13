package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	RepoOwner = "Zacy-Sokach"
	RepoName  = "PolyAgent"
	Repo      = RepoOwner + "/" + RepoName
)

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type Checker struct {
	client *http.Client
}

func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Checker) GetLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest version: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return release.TagName, nil
}

func (c *Checker) CheckForUpdate(currentVersion string) (bool, string, error) {
	latestVersion, err := c.GetLatestVersion()
	if err != nil {
		return false, "", err
	}
	
	if compareVersions(currentVersion, latestVersion) < 0 {
		return true, latestVersion, nil
	}
	
	return false, latestVersion, nil
}

func (c *Checker) GetDownloadURL(version string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	
	if os == "darwin" {
		os = "darwin"
	} else if os == "linux" {
		os = "linux"
	} else if os == "windows" {
		os = "windows"
	}
	
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}
	
	binaryName := fmt.Sprintf("polyagent-%s-%s", os, arch)
	if os == "windows" {
		binaryName += ".exe"
	}
	
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repo, version, binaryName)
}

func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		var p1, p2 int
		fmt.Sscanf(parts1[i], "%d", &p1)
		fmt.Sscanf(parts2[i], "%d", &p2)
		
		if p1 < p2 {
			return -1
		}
		if p1 > p2 {
			return 1
		}
	}
	
	if len(parts1) < len(parts2) {
		return -1
	}
	if len(parts1) > len(parts2) {
		return 1
	}
	
	return 0
}
