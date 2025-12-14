package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// dirItem 表示目录项信息
type dirItem struct {
	path  string
	info  os.FileInfo
	depth int
}

// GetCurrentDirContext 获取当前目录的上下文信息，包括目录结构和代码文件
// 添加了深度限制（最大5层）和权限检查，避免遍历过深或访问无权限的目录
// 优化：使用并发处理提高大目录遍历性能
func GetCurrentDirContext() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %w", err)
	}

	var sb strings.Builder
	sb.Grow(4096) // 预分配容量
	sb.WriteString(fmt.Sprintf("当前工作目录: %s\n\n", cwd))
	sb.WriteString("目录结构（最多显示5层深度）:\n")

	const maxDepth = 5
	const maxWorkers = 8 // 并发worker数量
	visitedSymlinks := make(map[string]bool)
	
	itemsChan := make(chan dirItem, 1000)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 启动worker池
	semaphore := make(chan struct{}, maxWorkers)
	
	// 收集根目录下的直接子项
	rootEntries, err := os.ReadDir(cwd)
	if err != nil {
		return "", fmt.Errorf("读取根目录失败: %w", err)
	}
	
	// 处理根目录下的直接子项
	for _, entry := range rootEntries {
		info, err := entry.Info()
		if err != nil {
			continue // 跳过错误
		}
		
		path := filepath.Join(cwd, entry.Name())
		depth := 0
		
		// 检查符号链接循环
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				continue
			}
			absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(path), target))
			if err != nil {
				continue
			}
			mu.Lock()
			if visitedSymlinks[absTarget] {
				mu.Unlock()
				continue
			}
			visitedSymlinks[absTarget] = true
			mu.Unlock()
		}
		
		// 处理目录项
		wg.Add(1)
		go func(p string, i os.FileInfo, d int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			processDirectoryItem(p, i, d, cwd, maxDepth, itemsChan, visitedSymlinks, &mu)
		}(path, info, depth)
	}
	
	// 等待所有处理完成
	go func() {
		wg.Wait()
		close(itemsChan)
	}()
	
	// 收集并排序结果
	var items []dirItem
	for item := range itemsChan {
		items = append(items, item)
	}
	
	// 按路径排序，确保输出一致性
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].path > items[j].path {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	
	// 输出结果
	for _, item := range items {
		indent := strings.Repeat("  ", item.depth)
		
		if item.info.IsDir() {
			sb.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, item.info.Name()))
		} else {
			ext := filepath.Ext(item.info.Name())
			if isCodeFile(ext) {
				sb.WriteString(fmt.Sprintf("%s📄 %s\n", indent, item.info.Name()))
			}
		}
	}

	return sb.String(), nil
}

// processDirectoryItem 处理单个目录项
func processDirectoryItem(path string, info os.FileInfo, depth int, cwd string, maxDepth int, itemsChan chan dirItem, visitedSymlinks map[string]bool, mu *sync.Mutex) {
	relPath, _ := filepath.Rel(cwd, path)
	if relPath == "." {
		return
	}
	
	// 检查深度限制
	if depth > maxDepth {
		return
	}
	
	// 发送当前项到通道
	itemsChan <- dirItem{path, info, depth}
	
	// 如果是目录且未达到最大深度，递归处理子项
	if info.IsDir() && depth < maxDepth {
		entries, err := os.ReadDir(path)
		if err != nil {
			return // 跳过无法读取的目录
		}
		
		for _, entry := range entries {
			childInfo, err := entry.Info()
			if err != nil {
				continue
			}
			
			childPath := filepath.Join(path, entry.Name())
			
			// 检查符号链接循环
			if childInfo.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(childPath)
				if err != nil {
					continue
				}
				absTarget, err := filepath.Abs(filepath.Join(filepath.Dir(childPath), target))
				if err != nil {
					continue
				}
				mu.Lock()
				if visitedSymlinks[absTarget] {
					mu.Unlock()
					continue
				}
				visitedSymlinks[absTarget] = true
				mu.Unlock()
			}
			
			// 递归处理子项
			processDirectoryItem(childPath, childInfo, depth+1, cwd, maxDepth, itemsChan, visitedSymlinks, mu)
		}
	}
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
