package utils

import (
	"os"
	"path/filepath"
)

// GetConfigDir 获取跨平台的配置目录
// Windows: %APPDATA%/polyagent
// Linux/macOS: ~/.config/polyagent
func GetConfigDir() (string, error) {
	// 检查是否设置了自定义配置目录
	if configHome := os.Getenv("POLYAGENT_CONFIG_HOME"); configHome != "" {
		return configHome, nil
	}

	// Windows: 使用 APPDATA
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "polyagent"), nil
	}

	// Linux/macOS: 使用 XDG_CONFIG_HOME 或 ~/.config
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, "polyagent"), nil
	}

	// 默认使用用户主目录下的 .config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "polyagent"), nil
}

// GetConfigPathForDisplay 获取用于显示的配置路径字符串
func GetConfigPathForDisplay() string {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return filepath.Join(appData, "polyagent", "config.yaml") + " (Windows)"
	}
	return "~/.config/polyagent/config.yaml (Linux/macOS)"
}
