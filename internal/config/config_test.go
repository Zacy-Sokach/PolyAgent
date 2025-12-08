package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStruct(t *testing.T) {
	// 测试Config结构体
	config := &Config{
		APIKey: "test-key",
		Model:  "glm-4.5",
	}

	if config.APIKey != "test-key" {
		t.Errorf("APIKey not set correctly: %s", config.APIKey)
	}
	if config.Model != "glm-4.5" {
		t.Errorf("Model not set correctly: %s", config.Model)
	}
}

func TestGetConfigPath(t *testing.T) {
	// 测试获取配置路径
	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath failed: %v", err)
	}

	// 检查路径格式
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedPath := filepath.Join(homeDir, ".config", "polyagent", "config.yaml")
	if path != expectedPath {
		t.Errorf("Config path mismatch: got %s, want %s", path, expectedPath)
	}
}

func TestSaveAndLoadConfigIntegration(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 临时修改HOME环境变量
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// 测试保存配置
	testConfig := &Config{
		APIKey: "test-api-key-123",
		Model:  "glm-4.5",
	}

	err := SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// 验证配置文件已创建
	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// 测试加载配置
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedConfig.APIKey != testConfig.APIKey {
		t.Errorf("Loaded APIKey %q doesn't match saved %q", loadedConfig.APIKey, testConfig.APIKey)
	}
	if loadedConfig.Model != testConfig.Model {
		t.Errorf("Loaded Model %q doesn't match saved %q", loadedConfig.Model, testConfig.Model)
	}
}

func TestLoadConfigWhenNotExists(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 临时修改HOME环境变量
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// 确保配置文件不存在
	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath failed: %v", err)
	}
	
	// 删除可能存在的配置文件
	os.Remove(configPath)

	// 加载不存在的配置应该返回默认值
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed when config doesn't exist: %v", err)
	}

	if config.APIKey != "" {
		t.Errorf("Expected empty APIKey for new config, got %q", config.APIKey)
	}
	if config.Model != "glm-4.5" {
		t.Errorf("Expected default model 'glm-4.5', got %q", config.Model)
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 临时修改HOME环境变量
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// 创建无效的配置文件
	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath failed: %v", err)
	}

	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte("invalid: yaml: content: [}"), 0644)

	// 加载无效配置应该返回错误
	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}