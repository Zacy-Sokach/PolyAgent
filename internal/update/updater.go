package update

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type Updater struct {
	checker *Checker
	client  *http.Client
}

func NewUpdater() *Updater {
	return &Updater{
		checker: NewChecker(),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (u *Updater) Update(currentVersion string) error {
	hasUpdate, latestVersion, err := u.checker.CheckForUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for update: %w", err)
	}
	
	if !hasUpdate {
		return fmt.Errorf("already running the latest version (%s)", currentVersion)
	}
	
	fmt.Printf("Updating from %s to %s...\n", currentVersion, latestVersion)
	
	downloadURL := u.checker.GetDownloadURL(latestVersion)
	checksumURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/checksums.txt", Repo, latestVersion)
	
	tempDir, err := os.MkdirTemp("", "polyagent-update-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	binaryPath := filepath.Join(tempDir, "polyagent")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	
	if err := u.downloadFile(downloadURL, binaryPath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	
	if err := u.verifyChecksum(binaryPath, checksumURL); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}
	
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}
	
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}
	
	backupPath := executablePath + ".backup"
	if err := os.Rename(executablePath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	if err := os.Rename(binaryPath, executablePath); err != nil {
		os.Rename(backupPath, executablePath)
		return fmt.Errorf("failed to install update: %w", err)
	}
	
	os.Remove(backupPath)
	
	fmt.Printf("Successfully updated to %s!\n", latestVersion)
	
	return nil
}

func (u *Updater) downloadFile(url, destPath string) error {
	resp, err := u.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	return err
}

func (u *Updater) verifyChecksum(filePath, checksumURL string) error {
	resp, err := u.client.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum file not found")
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksums: %w", err)
	}
	
	fileName := filepath.Base(filePath)
	lines := string(body)
	for _, line := range splitLines(lines) {
		parts := splitFields(line)
		if len(parts) >= 2 && parts[1] == fileName {
			expectedChecksum := parts[0]
			actualChecksum, err := calculateSHA256(filePath)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum: %w", err)
			}
			
			if expectedChecksum != actualChecksum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("checksum not found for %s", fileName)
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitFields(s string) []string {
	var fields []string
	start := 0
	inField := false
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			if inField {
				fields = append(fields, s[start:i])
				inField = false
			}
		} else {
			if !inField {
				start = i
				inField = true
			}
		}
	}
	if inField {
		fields = append(fields, s[start:])
	}
	return fields
}

func (u *Updater) InstallFromURL(url string) error {
	fmt.Printf("Installing PolyAgent from %s...\n", url)
	
	tempDir, err := os.MkdirTemp("", "polyagent-install-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	binaryPath := filepath.Join(tempDir, "polyagent")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	
	if err := u.downloadFile(url, binaryPath); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	
	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}
	
	installDir := "/usr/local/bin"
	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "PolyAgent")
		os.MkdirAll(installDir, 0755)
	}
	
	destPath := filepath.Join(installDir, "polyagent")
	if runtime.GOOS == "windows" {
		destPath += ".exe"
	}
	
	if err := os.Rename(binaryPath, destPath); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}
	
	fmt.Printf("Successfully installed PolyAgent to %s\n", destPath)
	
	if runtime.GOOS != "windows" {
		fmt.Printf("You can now run 'polyagent' from anywhere!\n")
	} else {
		fmt.Printf("Please add %s to your PATH or restart your terminal.\n", installDir)
	}
	
	return nil
}

func (u *Updater) RunInstaller() error {
	var installerURL string
	if runtime.GOOS == "windows" {
		installerURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/main/scripts/install.ps1", Repo)
		cmd := exec.Command("powershell", "-Command", fmt.Sprintf("irm %s | iex", installerURL))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	} else {
		installerURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/main/scripts/install.sh", Repo)
		cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL %s | bash", installerURL))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
