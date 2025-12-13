package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ReadFileTool 读取文件工具（基于 FileEngine）
type ReadFileTool struct {
	engine *FileEngine
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read file content with caching support. Use force_refresh=true to skip cache."
}

func (t *ReadFileTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file",
			},
			"force_refresh": map[string]interface{}{
				"type":        "boolean",
				"description": "Skip cache and read from disk",
				"default":     false,
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	forceRefresh := false
	if fr, ok := args["force_refresh"].(bool); ok {
		forceRefresh = fr
	}

	content, err := t.engine.ReadFile(path, forceRefresh)
	if err != nil {
		return nil, ConvertToMCPError(err)
	}

	return string(content), nil
}

// WriteFileTool 写入文件工具（基于 FileEngine）
type WriteFileTool struct {
	engine *FileEngine
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to file with automatic backup. Creates backup before overwriting."
}

func (t *WriteFileTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write",
			},
			"backup": map[string]interface{}{
				"type":        "boolean",
				"description": "Create backup before writing",
				"default":     true,
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter: content")
	}

	backup := true
	if b, ok := args["backup"].(bool); ok {
		backup = b
	}

	err := t.engine.WriteFile(path, []byte(content), backup)
	if err != nil {
		return nil, ConvertToMCPError(err)
	}

	result := map[string]interface{}{
		"success": true,
		"path":    path,
	}

	if backup {
		result["backup_created"] = true
		result["backup_dir"] = t.engine.config.BackupDir
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

// ReplaceTool 替换文件内容工具（基于 FileEngine）
type ReplaceTool struct {
	engine *FileEngine
}

func (t *ReplaceTool) Name() string {
	return "replace"
}

func (t *ReplaceTool) Description() string {
	return "Replace text in file using string or regex matching. Creates backup before modification."
}

func (t *ReplaceTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute path to the file",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "String or regex pattern to replace",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "Replacement string",
			},
			"use_regex": map[string]interface{}{
				"type":        "boolean",
				"description": "Use regex pattern matching",
				"default":     false,
			},
			"backup": map[string]interface{}{
				"type":        "boolean",
				"description": "Create backup before modification",
				"default":     true,
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}
}

func (t *ReplaceTool) Execute(args map[string]interface{}) (interface{}, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("missing required parameter: file_path")
	}

	oldString, ok := args["old_string"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter: old_string")
	}

	newString, ok := args["new_string"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter: new_string")
	}

	useRegex := false
	if ur, ok := args["use_regex"].(bool); ok {
		useRegex = ur
	}

	backup := true
	if b, ok := args["backup"].(bool); ok {
		backup = b
	}

	// 读取文件内容
	content, err := t.engine.ReadFile(filePath, false)
	if err != nil {
		return nil, ConvertToMCPError(fmt.Errorf("failed to read file: %w", err))
	}

	// 执行替换
	var newContent string
	if useRegex {
		// 正则表达式替换
		re, err := regexp.Compile(oldString)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		newContent = re.ReplaceAllString(string(content), newString)
	} else {
		// 普通字符串替换
		newContent = strings.ReplaceAll(string(content), oldString, newString)
	}

	// 写入文件
	err = t.engine.WriteFile(filePath, []byte(newContent), backup)
	if err != nil {
		return nil, ConvertToMCPError(fmt.Errorf("failed to write file: %w", err))
	}

	result := map[string]interface{}{
		"success":     true,
		"file_path":   filePath,
		"replacements": strings.Count(string(content), oldString),
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

// DiagnoseFileTool 诊断文件工具
type DiagnoseFileTool struct {
	engine *FileEngine
}

func (t *DiagnoseFileTool) Name() string {
	return "diagnose_file"
}

func (t *DiagnoseFileTool) Description() string {
	return "Diagnose file access issues, check path validation, cache status, and provide suggestions."
}

func (t *DiagnoseFileTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to diagnose",
			},
		},
		"required": []string{"path"},
	}
}

func (t *DiagnoseFileTool) Execute(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("missing required parameter: path")
	}

	result := map[string]interface{}{
		"path":        path,
		"diagnosis":   make(map[string]interface{}),
		"suggestions": []string{},
	}

	diagnosis := result["diagnosis"].(map[string]interface{})

	// 1. 路径验证
	validationErr := t.engine.ValidatePath(path)
	if validationErr != nil {
		diagnosis["validation"] = map[string]interface{}{
			"allowed": false,
			"error":   validationErr.Error(),
		}
		result["suggestions"] = append(result["suggestions"].([]string), 
			"Check that the path is within your project directory")
	} else {
		diagnosis["validation"] = map[string]interface{}{
			"allowed": true,
		}
	}

	// 2. 文件信息
	if info, err := os.Stat(path); err == nil {
		diagnosis["file_info"] = map[string]interface{}{
			"exists":   true,
			"size":     info.Size(),
			"mode":     info.Mode().String(),
			"mod_time": info.ModTime().Format("2006-01-02 15:04:05"),
			"is_dir":   info.IsDir(),
		}

		// 检查文件大小
		if info.Size() > t.engine.config.MaxFileSize {
			result["suggestions"] = append(result["suggestions"].([]string),
				"File is too large, consider using offset and limit parameters")
		}
	} else {
		diagnosis["file_info"] = map[string]interface{}{
			"exists": false,
			"error":  err.Error(),
		}
		result["suggestions"] = append(result["suggestions"].([]string),
			"File does not exist, check the path")
	}

	// 3. 缓存状态
	if t.engine.cache != nil && t.engine.config.EnableCache {
		if _, hit := t.engine.cache.get(path); hit {
			diagnosis["cache_status"] = map[string]interface{}{
				"cached": true,
			}
			result["suggestions"] = append(result["suggestions"].([]string),
				"File is cached, use force_refresh=true to read from disk")
		} else {
			diagnosis["cache_status"] = map[string]interface{}{
				"cached": false,
			}
		}
	}

	// 4. 备份信息
	backupDir := t.engine.config.BackupDir
	if info, err := os.Stat(backupDir); err == nil && info.IsDir() {
		diagnosis["backup_info"] = map[string]interface{}{
			"backup_enabled": true,
			"backup_dir":     backupDir,
		}
	} else {
		diagnosis["backup_info"] = map[string]interface{}{
			"backup_enabled": false,
		}
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
