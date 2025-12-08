package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ToolHandler 工具处理器接口
type ToolHandler interface {
	Name() string
	Description() string
	GetSchema() map[string]interface{}
	Execute(args map[string]interface{}) (interface{}, error)
}

// 安全配置
var (
	// 允许访问的根目录（默认为当前工作目录）
	safeRootDir string
	// 是否启用严格路径检查
	strictPathCheck = true
)

// init 初始化安全配置
func init() {
	// 设置安全根目录为当前工作目录
	if wd, err := os.Getwd(); err == nil {
		safeRootDir = wd
	}
}

// validateSafePath 验证路径是否安全
func validateSafePath(path string) error {
	if !strictPathCheck {
		return nil // 如果禁用严格检查，则允许所有路径
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 转换为绝对路径
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("无法获取绝对路径: %w", err)
	}

	// 检查是否在安全根目录内
	relPath, err := filepath.Rel(safeRootDir, absPath)
	if err != nil {
		return fmt.Errorf("路径不在安全范围内: %w", err)
	}

	// 检查是否尝试访问上级目录
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("禁止访问上级目录: %s", path)
	}

	// 检查是否包含危险符号
	if strings.Contains(relPath, "~") || strings.Contains(relPath, "$") {
		return fmt.Errorf("路径包含危险符号: %s", path)
	}

	return nil
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]ToolHandler
}

// NewToolRegistry 创建新的工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolHandler),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool ToolHandler) {
	r.tools[tool.Name()] = tool
}

// GetTool 获取工具
func (r *ToolRegistry) GetTool(name string) (ToolHandler, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// ListTools 列出所有工具
func (r *ToolRegistry) ListTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, handler := range r.tools {
		tools = append(tools, Tool{
			Name:        handler.Name(),
			Description: handler.Description(),
		})
	}
	return tools
}

// HandleCallTool 处理工具调用
func (r *ToolRegistry) HandleCallTool(req CallToolRequest) (*CallToolResult, error) {
	// 添加恢复机制防止panic
	defer func() {
		if r := recover(); r != nil {
			// fmt.Printf("[MCP] HandleCallTool 恢复panic: %v\n", r)
		}
	}()

	handler, ok := r.GetTool(req.Name)
	if !ok {
		return nil, fmt.Errorf("工具未找到: %s", req.Name)
	}

	// 记录工具调用（用于调试）
	// argsJSON, _ := json.Marshal(req.Arguments)
	// fmt.Printf("[MCP] 调用工具: %s, 参数: %s\n", req.Name, string(argsJSON))

	// 检查参数是否为空
	if req.Arguments == nil {
		req.Arguments = make(map[string]interface{})
	}

	// 执行工具调用（添加错误恢复）
	result, err := func() (interface{}, error) {
		defer func() {
			if r := recover(); r != nil {
				// fmt.Printf("[MCP] 工具执行恢复panic: %s, 错误: %v\n", req.Name, r)
			}
		}()
		return handler.Execute(req.Arguments)
	}()

	if err != nil {
		// 记录详细错误信息
		// fmt.Printf("[MCP] 工具执行失败: %s, 错误: %v\n", req.Name, err)
		return nil, fmt.Errorf("工具执行失败: %w", err)
	}

	// 将结果转换为ToolResultContent
	content := ToolResultContent{
		Type: "text",
		Text: fmt.Sprintf("%v", result),
	}

	// fmt.Printf("[MCP] 工具执行成功: %s\n", req.Name)
	return &CallToolResult{
		Content: []ToolResultContent{content},
	}, nil
}

// 基础工具实现

// ReadFileTool 读取文件工具
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string                      { return "read_file" }
func (t *ReadFileTool) Description() string               { return "读取文件内容" }
func (t *ReadFileTool) GetSchema() map[string]interface{} { return ReadFileSchema }

func (t *ReadFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	// 安全验证：检查路径是否在允许的范围内
	if err := validateSafePath(path); err != nil {
		return nil, fmt.Errorf("路径安全验证失败: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return string(content), nil
}

// WriteFileTool 写入文件工具
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string                      { return "write_file" }
func (t *WriteFileTool) Description() string               { return "写入文件内容" }
func (t *WriteFileTool) GetSchema() map[string]interface{} { return WriteFileSchema }

func (t *WriteFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的content参数")
	}

	// 安全验证：检查路径是否在允许的范围内
	if err := validateSafePath(path); err != nil {
		return nil, fmt.Errorf("路径安全验证失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return "文件写入成功", nil
}

// ListDirectoryTool 列出目录工具
type ListDirectoryTool struct{}

func (t *ListDirectoryTool) Name() string                      { return "list_directory" }
func (t *ListDirectoryTool) Description() string               { return "列出目录内容" }
func (t *ListDirectoryTool) GetSchema() map[string]interface{} { return ListDirectorySchema }

func (t *ListDirectoryTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	var result []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		result = append(result, name)
	}

	return strings.Join(result, "\n"), nil
}

// SearchFileContentTool 搜索文件内容工具
type SearchFileContentTool struct{}

func (t *SearchFileContentTool) Name() string                      { return "search_file_content" }
func (t *SearchFileContentTool) Description() string               { return "在文件中搜索内容" }
func (t *SearchFileContentTool) GetSchema() map[string]interface{} { return SearchFileContentSchema }

func (t *SearchFileContentTool) Execute(args map[string]interface{}) (interface{}, error) {
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

	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式: %w", err)
	}

	var results []string

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

		// 性能优化：检查文件大小，避免读取过大文件
		if info.Size() > 10*1024*1024 { // 10MB限制
			return nil // 跳过大于10MB的文件
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // 跳过无法读取的文件
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", filePath, i+1, line))

				// 限制每个文件的最大匹配数
				if len(results) >= 1000 {
					return fmt.Errorf("达到最大匹配数限制: 1000")
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("搜索文件时出错: %w", err)
	}

	if len(results) == 0 {
		return "未找到匹配的内容", nil
	}

	return strings.Join(results, "\n"), nil
}

// GlobTool 文件匹配工具
type GlobTool struct{}

func (t *GlobTool) Name() string                      { return "glob" }
func (t *GlobTool) Description() string               { return "使用glob模式匹配文件" }
func (t *GlobTool) GetSchema() map[string]interface{} { return GlobSchema }

func (t *GlobTool) Execute(args map[string]interface{}) (interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的pattern参数")
	}

	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	matches, err := filepath.Glob(filepath.Join(path, pattern))
	if err != nil {
		return nil, fmt.Errorf("glob匹配失败: %w", err)
	}

	if len(matches) == 0 {
		return "未找到匹配的文件", nil
	}

	return strings.Join(matches, "\n"), nil
}

// ReplaceTool 替换文件内容工具
type ReplaceTool struct{}

func (t *ReplaceTool) Name() string                      { return "replace" }
func (t *ReplaceTool) Description() string               { return "替换文件中的内容" }
func (t *ReplaceTool) GetSchema() map[string]interface{} { return ReplaceSchema }

func (t *ReplaceTool) Execute(args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的file_path参数")
	}

	oldString, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的old_string参数")
	}

	newString, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的new_string参数")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	newContent := strings.ReplaceAll(string(content), oldString, newString)

	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return "替换完成", nil
}

// RunShellCommandTool 执行shell命令工具
type RunShellCommandTool struct{}

func (t *RunShellCommandTool) Name() string                      { return "run_shell_command" }
func (t *RunShellCommandTool) Description() string               { return "执行shell命令" }
func (t *RunShellCommandTool) GetSchema() map[string]interface{} { return RunShellCommandSchema }

func (t *RunShellCommandTool) Execute(args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的command参数")
	}

	// 注意：这里简化实现，实际应该使用exec.Command
	// 由于安全考虑，这里只返回示例
	return fmt.Sprintf("执行命令: %s\n(实际实现需要使用exec.Command)", command), nil
}

// CreateFileTool 创建文件工具
type CreateFileTool struct{}

func (t *CreateFileTool) Name() string                      { return "create_file" }
func (t *CreateFileTool) Description() string               { return "创建新文件" }
func (t *CreateFileTool) GetSchema() map[string]interface{} { return CreateFileSchema }

func (t *CreateFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的content参数")
	}

	overwrite := false
	if ow, ok := args["overwrite"].(bool); ok {
		overwrite = ow
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); err == nil && !overwrite {
		return nil, fmt.Errorf("文件已存在，如需覆盖请设置overwrite=true")
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}

	return "文件创建成功", nil
}

// DeleteFileTool 删除文件工具
type DeleteFileTool struct{}

func (t *DeleteFileTool) Name() string                      { return "delete_file" }
func (t *DeleteFileTool) Description() string               { return "删除文件或目录" }
func (t *DeleteFileTool) GetSchema() map[string]interface{} { return DeleteFileSchema }

func (t *DeleteFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	recursive := false
	if rec, ok := args["recursive"].(bool); ok {
		recursive = rec
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	if info.IsDir() && !recursive {
		return nil, fmt.Errorf("目录非空，如需删除请设置recursive=true")
	}

	if info.IsDir() {
		if err := os.RemoveAll(path); err != nil {
			return nil, fmt.Errorf("删除目录失败: %w", err)
		}
	} else {
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("删除文件失败: %w", err)
		}
	}

	return "删除成功", nil
}

// MoveFileTool 移动文件工具
type MoveFileTool struct{}

func (t *MoveFileTool) Name() string                      { return "move_file" }
func (t *MoveFileTool) Description() string               { return "移动文件或目录" }
func (t *MoveFileTool) GetSchema() map[string]interface{} { return MoveFileSchema }

func (t *MoveFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	source, ok := args["source"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的source参数")
	}

	destination, ok := args["destination"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的destination参数")
	}

	overwrite := false
	if ow, ok := args["overwrite"].(bool); ok {
		overwrite = ow
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(destination); err == nil && !overwrite {
		return nil, fmt.Errorf("目标文件已存在，如需覆盖请设置overwrite=true")
	}

	if err := os.Rename(source, destination); err != nil {
		return nil, fmt.Errorf("移动文件失败: %w", err)
	}

	return "移动成功", nil
}

// CopyFileTool 复制文件工具
type CopyFileTool struct{}

func (t *CopyFileTool) Name() string                      { return "copy_file" }
func (t *CopyFileTool) Description() string               { return "复制文件或目录" }
func (t *CopyFileTool) GetSchema() map[string]interface{} { return CopyFileSchema }

func (t *CopyFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	source, ok := args["source"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的source参数")
	}

	destination, ok := args["destination"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的destination参数")
	}

	overwrite := false
	if ow, ok := args["overwrite"].(bool); ok {
		overwrite = ow
	}

	// 检查目标文件是否存在
	if _, err := os.Stat(destination); err == nil && !overwrite {
		return nil, fmt.Errorf("目标文件已存在，如需覆盖请设置overwrite=true")
	}

	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("读取源文件失败: %w", err)
	}

	// 确保目标目录存在
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(destination, sourceContent, 0644); err != nil {
		return nil, fmt.Errorf("写入目标文件失败: %w", err)
	}

	return "复制成功", nil
}

// GetFileInfoTool 获取文件信息工具
type GetFileInfoTool struct{}

func (t *GetFileInfoTool) Name() string                      { return "get_file_info" }
func (t *GetFileInfoTool) Description() string               { return "获取文件或目录信息" }
func (t *GetFileInfoTool) GetSchema() map[string]interface{} { return GetFileInfoSchema }

func (t *GetFileInfoTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的path参数")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	result := map[string]interface{}{
		"name":     info.Name(),
		"size":     info.Size(),
		"mode":     info.Mode().String(),
		"mod_time": info.ModTime().Format("2006-01-02 15:04:05"),
		"is_dir":   info.IsDir(),
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	return string(resultBytes), nil
}

// ExecuteCodeTool 执行代码工具
type ExecuteCodeTool struct{}

func (t *ExecuteCodeTool) Name() string                      { return "execute_code" }
func (t *ExecuteCodeTool) Description() string               { return "执行代码片段" }
func (t *ExecuteCodeTool) GetSchema() map[string]interface{} { return ExecuteCodeSchema }

func (t *ExecuteCodeTool) Execute(args map[string]interface{}) (interface{}, error) {
	language, ok := args["language"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的language参数")
	}

	code, ok := args["code"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的code参数")
	}

	// 注意：这里简化实现，实际应该根据语言执行代码
	// 由于安全考虑，这里只返回示例
	return fmt.Sprintf("执行 %s 代码:\n%s\n\n(实际实现需要根据语言调用相应的解释器/编译器)", language, code), nil
}

// GitOperationTool Git操作工具
type GitOperationTool struct{}

func (t *GitOperationTool) Name() string                      { return "git_operation" }
func (t *GitOperationTool) Description() string               { return "执行Git操作" }
func (t *GitOperationTool) GetSchema() map[string]interface{} { return GitOperationSchema }

func (t *GitOperationTool) Execute(args map[string]interface{}) (interface{}, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少或无效的operation参数")
	}

	// 注意：这里简化实现，实际应该调用git命令
	// 由于安全考虑，这里只返回示例
	return fmt.Sprintf("执行Git操作: %s\n(实际实现需要调用git命令)", operation), nil
}

// GetCurrentTimeTool 获取当前时间工具
type GetCurrentTimeTool struct{}

func (t *GetCurrentTimeTool) Name() string { return "get_current_time" }
func (t *GetCurrentTimeTool) Description() string {
	return "获取当前的系统时间，可指定输出格式"
}
func (t *GetCurrentTimeTool) GetSchema() map[string]interface{} { return GetCurrentTimeSchema }

func (t *GetCurrentTimeTool) Execute(args map[string]interface{}) (interface{}, error) {
	format, ok := args["format"].(string)
	if !ok || format == "" {
		format = time.RFC3339 // 默认使用标准格式
	}

	// 简单的格式映射，以提高 LLM 的可用性
	switch format {
	case "long":
		format = time.RFC1123
	case "short":
		format = "15:04:05" // HH:MM:SS
	default:
		// 保持原样
	}

	return time.Now().Format(format), nil
}

// DefaultToolRegistry 创建默认工具注册表
func DefaultToolRegistry() *ToolRegistry {
	registry := NewToolRegistry()

	// 注册基础工具
	registry.Register(&ReadFileTool{})
	registry.Register(&WriteFileTool{})
	registry.Register(&ListDirectoryTool{})
	registry.Register(&SearchFileContentTool{})
	registry.Register(&GlobTool{})
	registry.Register(&ReplaceTool{})
	registry.Register(&RunShellCommandTool{})
	registry.Register(&CreateFileTool{})
	registry.Register(&DeleteFileTool{})
	registry.Register(&MoveFileTool{})
	registry.Register(&CopyFileTool{})
	registry.Register(&GetFileInfoTool{})
	registry.Register(&ExecuteCodeTool{})
	registry.Register(&GitOperationTool{})
	registry.Register(&GetCurrentTimeTool{})

	// 注册 Tavily 搜索工具
	registry.Register(NewTavilySearchTool())
	registry.Register(NewTavilyCrawlTool())

	// 注册高级工具
	RegisterAdvancedTools(registry)

	return registry
}
