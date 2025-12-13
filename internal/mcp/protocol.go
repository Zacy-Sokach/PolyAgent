package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MCP协议版本
const ProtocolVersion = "2024-11-05"

// 错误码定义
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeToolError      = -32000
	
	// FileEngine 相关错误码
	CodePathNotAllowed = -32001
	CodeFileTooLarge   = -32002
	CodeFileNotFound   = -32003
	CodeBackupFailed   = -32004
	CodeCacheError     = -32005
	CodeReadError      = -32006
	CodeWriteError     = -32007
)

// ConvertToMCPError 将错误转换为 MCP 错误格式
func ConvertToMCPError(err error) *JSONRPCError {
	if err == nil {
		return nil
	}
	
	code := CodeInternalError
	data := map[string]interface{}{
		"original_error": err.Error(),
	}
	
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "outside allowed roots"):
		code = CodePathNotAllowed
		data["suggestion"] = "Check that the path is within your project directory"
		
	case strings.Contains(errStr, "file too large"):
		code = CodeFileTooLarge
		data["max_size_mb"] = 10
		data["suggestion"] = "Try reading a portion of the file using offset and limit"
		
	case strings.Contains(errStr, "no such file") || strings.Contains(errStr, "file does not exist"):
		code = CodeFileNotFound
		data["suggestion"] = "Verify the file path exists"
		
	case strings.Contains(errStr, "backup failed"):
		code = CodeBackupFailed
		data["suggestion"] = "Check disk space and backup directory permissions"
		
	case strings.Contains(errStr, "permission denied"):
		data["suggestion"] = "Check file permissions"
		
	case strings.Contains(errStr, "file type not allowed"):
		code = CodePathNotAllowed
		data["suggestion"] = "The file extension is blacklisted for security reasons"
	}
	
	return &JSONRPCError{
		Code:    code,
		Message: err.Error(),
		Data:    data,
	}
}

// JSON-RPC消息类型
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error 实现 error 接口
func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("MCP Error %d: %s", e.Code, e.Message)
}

// MCP特定消息类型
type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      *ClientInfo        `json:"clientInfo,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      *ServerInfo        `json:"serverInfo,omitempty"`
}

type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// 工具相关类型
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ListToolsRequest struct{}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []ToolResultContent `json:"content"`
}

type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// 工具参数Schema定义
var (
	ReadFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件的绝对路径",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "起始行号（0-based）",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "读取行数限制",
			},
		},
		"required":             []string{"path"},
		"additionalProperties": false,
	}

	WriteFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件的绝对路径",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "文件内容",
			},
		},
		"required":             []string{"path", "content"},
		"additionalProperties": false,
	}

	ListDirectorySchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "目录的绝对路径",
			},
			"ignore": map[string]interface{}{
				"type":        "array",
				"description": "忽略的glob模式",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"path"},
	}

	SearchFileContentSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "正则表达式模式",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "搜索路径（文件或目录）",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "文件包含模式（glob）",
			},
		},
		"required": []string{"pattern"},
	}

	GlobSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "glob模式",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "搜索根目录",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "是否区分大小写",
			},
		},
		"required": []string{"pattern"},
	}

	ReplaceSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "文件的绝对路径",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "要替换的旧字符串或正则表达式模式",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "替换后的新字符串",
			},
			"use_regex": map[string]interface{}{
				"type":        "boolean",
				"description": "是否使用正则表达式进行替换",
				"default":     false,
			},
			"expected_replacements": map[string]interface{}{
				"type":        "integer",
				"description": "期望的替换次数",
			},
		},
		"required": []string{"file_path", "old_string", "new_string"},
	}

	RunShellCommandSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "要执行的shell命令",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "命令描述",
			},
			"dir_path": map[string]interface{}{
				"type":        "string",
				"description": "执行目录",
			},
		},
		"required": []string{"command"},
	}

	// 扩展工具Schema
	CreateFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件的绝对路径",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "文件内容",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "是否覆盖已存在的文件",
				"default":     false,
			},
		},
		"required": []string{"path", "content"},
	}

	DeleteFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件的绝对路径",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "是否递归删除目录",
				"default":     false,
			},
		},
		"required": []string{"path"},
	}

	MoveFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source": map[string]interface{}{
				"type":        "string",
				"description": "源文件/目录路径",
			},
			"destination": map[string]interface{}{
				"type":        "string",
				"description": "目标路径",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "是否覆盖已存在的文件",
				"default":     false,
			},
		},
		"required": []string{"source", "destination"},
	}

	CopyFileSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source": map[string]interface{}{
				"type":        "string",
				"description": "源文件/目录路径",
			},
			"destination": map[string]interface{}{
				"type":        "string",
				"description": "目标路径",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "是否覆盖已存在的文件",
				"default":     false,
			},
		},
		"required": []string{"source", "destination"},
	}

	GetFileInfoSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "文件/目录路径",
			},
		},
		"required": []string{"path"},
	}

	ExecuteCodeSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"language": map[string]interface{}{
				"type":        "string",
				"description": "编程语言（go, python, javascript等）",
				"enum":        []string{"go", "python", "javascript", "typescript", "bash", "shell"},
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "要执行的代码",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "超时时间（秒）",
				"default":     30,
			},
		},
		"required": []string{"language", "code"},
	}

	GitOperationSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Git操作",
				"enum":        []string{"status", "diff", "log", "add", "commit", "push", "pull", "branch", "checkout"},
			},
			"args": map[string]interface{}{
				"type":        "array",
				"description": "操作参数",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"operation"},
	}

	GetCurrentTimeSchema = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"format": map[string]interface{}{
				"type":        "string",
				"description": "可选的时间格式。例如: 'long' (RFC1123), 'short' (HH:MM:SS), 或 Go 标准库支持的布局字符串 (如 '2006-01-02')",
			},
		},
	}
)

// 错误码定义
// 创建错误响应
func NewError(code int, message string, data interface{}) *JSONRPCError {
	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// 创建成功响应
func NewResult(id json.RawMessage, result interface{}) (*JSONRPCResponse, error) {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultBytes,
	}, nil
}
