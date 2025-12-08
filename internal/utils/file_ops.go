package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func InsertCodeToFile(filePath, code string, lineNumber int) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	if lineNumber < 1 || lineNumber > len(lines)+1 {
		return fmt.Errorf("行号 %d 超出范围 (1-%d)", lineNumber, len(lines)+1)
	}

	newLines := make([]string, 0, len(lines)+strings.Count(code, "\n")+1)
	newLines = append(newLines, lines[:lineNumber-1]...)
	newLines = append(newLines, strings.Split(code, "\n")...)
	newLines = append(newLines, lines[lineNumber-1:]...)

	newContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func AppendCodeToFile(filePath, code string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	newContent := string(content)
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += code + "\n"

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func CreateNewFile(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	return nil
}

func GetCurrentFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	files, err := os.ReadDir(cwd)
	if err != nil {
		return "", fmt.Errorf("读取目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && isCodeFile(filepath.Ext(file.Name())) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("未找到代码文件")
}
