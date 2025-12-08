package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsCodeFile(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".go", true},
		{".py", true},
		{".js", true},
		{".ts", true},
		{".java", true},
		{".md", true},
		{".json", true},
		{".txt", false},
		{".exe", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isCodeFile(tt.ext)
		if result != tt.expected {
			t.Errorf("isCodeFile(%q) = %v, want %v", tt.ext, result, tt.expected)
		}
	}
}

func TestGetFileContent(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "Hello, World!"

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 测试读取文件
	result, err := GetFileContent(tmpFile)
	if err != nil {
		t.Errorf("GetFileContent failed: %v", err)
	}
	if result != content {
		t.Errorf("GetFileContent returned %q, want %q", result, content)
	}

	// 测试不存在的文件
	_, err = GetFileContent("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetCurrentDirContext(t *testing.T) {
	// 创建测试目录结构
	tmpDir := t.TempDir()

	// 创建一些测试文件和目录
	os.MkdirAll(filepath.Join(tmpDir, "subdir1", "subdir2"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir1", "test.py"), []byte("print('hello')"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir1", "subdir2", "data.txt"), []byte("data"), 0644)

	// 保存当前目录并切换到测试目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// 测试获取目录上下文
	result, err := GetCurrentDirContext()
	if err != nil {
		t.Errorf("GetCurrentDirContext failed: %v", err)
	}

	// 检查结果包含预期内容
	if len(result) == 0 {
		t.Error("GetCurrentDirContext returned empty result")
	}

	// 应该包含目录信息
	if !contains(result, "当前工作目录:") {
		t.Error("Result should contain directory information")
	}

	// 应该包含代码文件
	if !contains(result, "main.go") {
		t.Error("Result should contain main.go")
	}
}

func TestGetCurrentFileContext(t *testing.T) {
	// 创建测试目录
	tmpDir := t.TempDir()

	// 创建测试文件
	os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package test\n\nfunc Test() {}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte("data"), 0644) // 非代码文件

	// 保存当前目录并切换到测试目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// 测试获取文件上下文
	result, err := GetCurrentFileContext()
	if err != nil {
		t.Errorf("GetCurrentFileContext failed: %v", err)
	}

	// 检查结果
	if len(result) == 0 {
		t.Error("GetCurrentFileContext returned empty result")
	}

	// 应该包含代码文件内容
	if !contains(result, "test.go") {
		t.Error("Result should contain test.go")
	}
	if !contains(result, "package test") {
		t.Error("Result should contain file content")
	}

	// 不应该包含非代码文件
	if contains(result, "data.txt") {
		t.Error("Result should not contain non-code files")
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
