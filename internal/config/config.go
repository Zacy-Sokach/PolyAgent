package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey       string `yaml:"api_key"`
	Model        string `yaml:"model"`
	TavilyAPIKey string `yaml:"tavily_api_key"`
}

func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Model: "glm-4.5",
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if config.Model == "" {
		config.Model = "glm-4.5"
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetTavilyAPIKey 获取 Tavily API Key
func GetTavilyAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.TavilyAPIKey, nil
}

// SaveTavilyAPIKey 保存 Tavily API Key
func SaveTavilyAPIKey(key string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	config.TavilyAPIKey = key
	return SaveConfig(config)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(homeDir, ".config", "polyagent", "config.yaml"), nil
}
