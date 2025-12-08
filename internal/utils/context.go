package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentDirContext 获取当前目录的上下文信息，包括目录结构和代码文件
// 添加了深度限制（最大5层）和权限检查，避免遍历过深或访问无权限的目录
func GetCurrentDirContext() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("当前工作目录: %s\n\n", cwd))
	sb.WriteString("目录结构（最多显示5层深度）:\n")

	const maxDepth = 5
	visitedSymlinks := make(map[string]bool)

	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 跳过无权限访问的目录
			if os.IsPermission(err) {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(cwd, path)
		if relPath == "." {
			return nil
		}

		// 检查深度限制
		depth := strings.Count(relPath, string(os.PathSeparator))
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查符号链接循环
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return nil // 跳过无法读取的符号链接
			}
			absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), target))
			if err != nil {
				return nil
			}
			if visitedSymlinks[absTarget] {
				return nil // 跳过已访问的符号链接
			}
			visitedSymlinks[absTarget] = true
		}

		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			sb.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, info.Name()))
		} else {
			ext := filepath.Ext(info.Name())
			if isCodeFile(ext) {
				sb.WriteString(fmt.Sprintf("%s📄 %s\n", indent, info.Name()))
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("遍历目录失败: %w", err)
	}

	return sb.String(), nil
}

// GetFileContent 读取指定文件的内容
// filePath: 文件路径
// 返回文件内容字符串，如果读取失败则返回错误
func GetFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(content), nil
}

// GetCurrentFileContext 获取当前目录下所有代码文件的内容
// 用于为AI提供代码上下文，只读取代码文件（根据扩展名判断）
// 返回包含所有代码文件内容的格式化字符串
func GetCurrentFileContext() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("当前目录: %s\n\n", cwd))

	files, err := os.ReadDir(cwd)
	if err != nil {
		return "", fmt.Errorf("读取目录失败: %w", err)
	}

	sb.WriteString("当前目录下的代码文件:\n")
	for _, file := range files {
		if !file.IsDir() && isCodeFile(filepath.Ext(file.Name())) {
			content, err := GetFileContent(file.Name())
			if err == nil {
				sb.WriteString(fmt.Sprintf("\n=== %s ===\n", file.Name()))
				sb.WriteString(content)
				sb.WriteString("\n")
			}
		}
	}

	return sb.String(), nil
}

// isCodeFile 判断文件扩展名是否为代码文件
// ext: 文件扩展名（如 ".go", ".py"）
// 返回true如果是支持的代码文件类型
func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".java": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".rs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
		".md": true, ".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".html": true, ".css": true, ".scss": true, ".sql": true, ".sh": true,
	}
	return codeExts[ext]
}
