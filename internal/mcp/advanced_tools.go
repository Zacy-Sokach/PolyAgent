package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// AdvancedSearchTool 高级文件搜索工具（支持多文件、正则表达式、上下文）
type AdvancedSearchTool struct{}

func (t *AdvancedSearchTool) Name() string { return "advanced_search" }
func (t *AdvancedSearchTool) Description() string {
	return "高级文件搜索，支持正则表达式、上下文显示和多种过滤选项"
}
func (t *AdvancedSearchTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "正则表达式模式",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "搜索路径",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "文件包含模式",
			},
			"exclude": map[string]interface{}{
				"type":        "string",
				"description": "文件排除模式",
			},
			"context_lines": map[string]interface{}{
				"type":        "integer",
				"description": "上下文行数",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "最大结果数",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "是否区分大小写",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *AdvancedSearchTool) Execute(args map[string]interface{}) (interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的pattern参数")
	}

	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	include := "*"
	if inc, ok := args["include"].(string); ok && inc != "" {
		include = inc
	}

	exclude := ""
	if exc, ok := args["exclude"].(string); ok && exc != "" {
		exclude = exc
	}

	contextLines := 0
	if ctx, ok := args["context_lines"].(float64); ok {
		contextLines = int(ctx)
	}

	maxResults := 100
	if max, ok := args["max_results"].(float64); ok {
		maxResults = int(max)
	}

	caseSensitive := true
	if cs, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	// 编译正则表达式
	var re *regexp.Regexp
	var err error
	if caseSensitive {
		re, err = regexp.Compile(pattern)
	} else {
		re, err = regexp.Compile("(?i)" + pattern)
	}
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	var results []string
	resultCount := 0

	// 递归搜索文件
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件是否匹配include模式
		matched, err := filepath.Match(include, filepath.Base(filePath))
		if err != nil || !matched {
			return nil
		}

		// 检查文件是否匹配exclude模式
		if exclude != "" {
			excluded, err := filepath.Match(exclude, filepath.Base(filePath))
			if err == nil && excluded {
				return nil
			}
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				resultCount++
				if resultCount > maxResults {
					return fmt.Errorf("达到最大结果数限制: %d", maxResults)
				}

				// 添加上下文
				start := max(0, i-contextLines)
				end := min(len(lines)-1, i+contextLines)

				var context []string
				for j := start; j <= end; j++ {
					prefix := "  "
					if j == i {
						prefix = "> "
					}
					context = append(context, fmt.Sprintf("%s%4d: %s", prefix, j+1, lines[j]))
				}

				result := fmt.Sprintf("%s:%d\n%s", filePath, i+1, strings.Join(context, "\n"))
				results = append(results, result)
			}
		}

		return nil
	})

	if err != nil && !strings.Contains(err.Error(), "达到最大结果数限制") {
		return nil, fmt.Errorf("搜索文件时出错: %w", err)
	}

	if len(results) == 0 {
		return "未找到匹配的内容", nil
	}

	summary := fmt.Sprintf("找到 %d 个匹配项:\n\n", len(results))
	return summary + strings.Join(results, "\n\n"), nil
}

// GlobalReplaceTool 全局替换工具
type GlobalReplaceTool struct{}

func (t *GlobalReplaceTool) Name() string        { return "global_replace" }
func (t *GlobalReplaceTool) Description() string { return "在多个文件中全局替换内容" }
func (t *GlobalReplaceTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"old_pattern": map[string]interface{}{
				"type":        "string",
				"description": "要替换的正则表达式模式",
			},
			"new_pattern": map[string]interface{}{
				"type":        "string",
				"description": "替换后的模式",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "搜索路径",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "文件包含模式",
			},
			"exclude": map[string]interface{}{
				"type":        "string",
				"description": "文件排除模式",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "是否只预览不实际修改",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "是否区分大小写",
			},
		},
		"required": []string{"old_pattern", "new_pattern"},
	}
}

func (t *GlobalReplaceTool) Execute(args map[string]interface{}) (interface{}, error) {
	oldPattern, ok := args["old_pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的old_pattern参数")
	}

	newPattern, ok := args["new_pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的new_pattern参数")
	}

	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	include := "*"
	if inc, ok := args["include"].(string); ok && inc != "" {
		include = inc
	}

	exclude := ""
	if exc, ok := args["exclude"].(string); ok && exc != "" {
		exclude = exc
	}

	dryRun := false
	if dr, ok := args["dry_run"].(bool); ok {
		dryRun = dr
	}

	caseSensitive := true
	if cs, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	// 编译正则表达式
	var re *regexp.Regexp
	var err error
	if caseSensitive {
		re, err = regexp.Compile(oldPattern)
	} else {
		re, err = regexp.Compile("(?i)" + oldPattern)
	}
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	type fileChange struct {
		path    string
		changes []string
	}

	var changes []fileChange
	totalReplacements := 0

	// 递归搜索并替换文件
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件是否匹配include模式
		matched, err := filepath.Match(include, filepath.Base(filePath))
		if err != nil || !matched {
			return nil
		}

		// 检查文件是否匹配exclude模式
		if exclude != "" {
			excluded, err := filepath.Match(exclude, filepath.Base(filePath))
			if err == nil && excluded {
				return nil
			}
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		oldContent := string(content)
		newContent := re.ReplaceAllString(oldContent, newPattern)

		if oldContent != newContent {
			// 统计替换次数
			oldLines := strings.Split(oldContent, "\n")
			newLines := strings.Split(newContent, "\n")

			var fileChanges []string
			for i := 0; i < len(oldLines) && i < len(newLines); i++ {
				if oldLines[i] != newLines[i] {
					fileChanges = append(fileChanges, fmt.Sprintf("  L%d: %s → %s", i+1, oldLines[i], newLines[i]))
					totalReplacements++
				}
			}

			changes = append(changes, fileChange{
				path:    filePath,
				changes: fileChanges,
			})

			// 如果不是dry run，则实际写入文件
			if !dryRun {
				if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
					return fmt.Errorf("写入文件失败 %s: %w", filePath, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("全局替换时出错: %w", err)
	}

	// 生成报告
	var report strings.Builder
	if dryRun {
		report.WriteString("DRY RUN - 不会实际修改文件\n\n")
	}

	report.WriteString(fmt.Sprintf("在 %d 个文件中找到 %d 处替换:\n\n", len(changes), totalReplacements))

	for _, change := range changes {
		report.WriteString(fmt.Sprintf("%s:\n", change.path))
		for _, lineChange := range change.changes {
			report.WriteString(lineChange + "\n")
		}
		report.WriteString("\n")
	}

	if !dryRun {
		report.WriteString("所有替换已完成。")
	} else {
		report.WriteString("这是预览，要实际执行替换请设置 dry_run=false。")
	}

	return report.String(), nil
}

// FindAndReplaceTool 查找并替换工具（交互式）
type FindAndReplaceTool struct{}

func (t *FindAndReplaceTool) Name() string { return "find_and_replace" }
func (t *FindAndReplaceTool) Description() string {
	return "查找内容并提供交互式替换选项"
}
func (t *FindAndReplaceTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"search_pattern": map[string]interface{}{
				"type":        "string",
				"description": "搜索的正则表达式模式",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "搜索路径",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "文件包含模式",
			},
		},
		"required": []string{"search_pattern"},
	}
}

func (t *FindAndReplaceTool) Execute(args map[string]interface{}) (interface{}, error) {
	searchPattern, ok := args["search_pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的search_pattern参数")
	}

	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	include := "*"
	if inc, ok := args["include"].(string); ok && inc != "" {
		include = inc
	}

	// 编译正则表达式
	re, err := regexp.Compile(searchPattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	type match struct {
		file    string
		line    int
		content string
	}

	var matches []match

	// 查找所有匹配项
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件是否匹配include模式
		matched, err := filepath.Match(include, filepath.Base(filePath))
		if err != nil || !matched {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				matches = append(matches, match{
					file:    filePath,
					line:    i + 1,
					content: line,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("查找文件时出错: %w", err)
	}

	if len(matches) == 0 {
		return "未找到匹配的内容", nil
	}

	// 生成查找结果
	var result strings.Builder
	result.WriteString(fmt.Sprintf("找到 %d 个匹配项:\n\n", len(matches)))

	for i, match := range matches {
		result.WriteString(fmt.Sprintf("%d. %s:%d\n", i+1, match.file, match.line))
		result.WriteString(fmt.Sprintf("   %s\n\n", match.content))
	}

	result.WriteString("要替换这些匹配项，请使用 global_replace 工具。")
	return result.String(), nil
}

// FileStatsTool 文件统计工具
type FileStatsTool struct{}

func (t *FileStatsTool) Name() string { return "file_stats" }
func (t *FileStatsTool) Description() string {
	return "获取文件统计信息（行数、大小、类型等）"
}
func (t *FileStatsTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "统计路径",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "文件包含模式",
			},
		},
	}
}

func (t *FileStatsTool) Execute(args map[string]interface{}) (interface{}, error) {
	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	include := "*"
	if inc, ok := args["include"].(string); ok && inc != "" {
		include = inc
	}

	type fileStat struct {
		name  string
		size  int64
		lines int
		ext   string
	}

	var stats []fileStat
	totalFiles := 0
	totalLines := 0
	totalSize := int64(0)

	// 收集统计信息
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件是否匹配include模式
		matched, err := filepath.Match(include, filepath.Base(filePath))
		if err != nil || !matched {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		lines := strings.Count(string(content), "\n")
		if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
			lines++
		}

		stats = append(stats, fileStat{
			name:  filePath,
			size:  info.Size(),
			lines: lines,
			ext:   strings.ToLower(filepath.Ext(filePath)),
		})

		totalFiles++
		totalLines += lines
		totalSize += info.Size()

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("收集文件统计信息时出错: %w", err)
	}

	// 按扩展名分组统计
	extStats := make(map[string]struct {
		count int
		lines int
		size  int64
	})

	for _, stat := range stats {
		ext := stat.ext
		if ext == "" {
			ext = "(无扩展名)"
		}

		es := extStats[ext]
		es.count++
		es.lines += stat.lines
		es.size += stat.size
		extStats[ext] = es
	}

	// 生成报告
	var report strings.Builder
	report.WriteString(fmt.Sprintf("文件统计报告:\n"))
	report.WriteString(fmt.Sprintf("总文件数: %d\n", totalFiles))
	report.WriteString(fmt.Sprintf("总行数: %d\n", totalLines))
	report.WriteString(fmt.Sprintf("总大小: %.2f KB\n\n", float64(totalSize)/1024))

	report.WriteString("按扩展名统计:\n")
	for ext, es := range extStats {
		report.WriteString(fmt.Sprintf("  %s: %d 个文件, %d 行, %.2f KB\n",
			ext, es.count, es.lines, float64(es.size)/1024))
	}

	return report.String(), nil
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RegisterAdvancedTools 注册高级工具
func RegisterAdvancedTools(registry *ToolRegistry) {
	registry.Register(&AdvancedSearchTool{})
	registry.Register(&GlobalReplaceTool{})
	registry.Register(&FindAndReplaceTool{})
	registry.Register(&FileStatsTool{})
	// WebSearchTool 已被 TavilySearchTool 和 TavilyCrawlTool 替代
}
