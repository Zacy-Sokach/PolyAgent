package mcp

import (
	"testing"
)

func TestToolRegistry(t *testing.T) {
	registry := DefaultToolRegistry()

	// 测试工具数量
	tools := registry.ListTools()
	if len(tools) == 0 {
		t.Error("工具注册表中没有工具")
	}

	// 测试常用工具是否存在
	requiredTools := []string{"read_file", "write_file", "get_current_time"}
	for _, toolName := range requiredTools {
		if _, exists := registry.GetTool(toolName); !exists {
			t.Errorf("缺少必需的工具: %s", toolName)
		}
	}

	// 测试时间工具
	if handler, exists := registry.GetTool("get_current_time"); exists {
		result, err := handler.Execute(map[string]interface{}{})
		if err != nil {
			t.Errorf("时间工具执行失败: %v", err)
		}
		if resultStr, ok := result.(string); !ok || resultStr == "" {
			t.Error("时间工具返回无效结果")
		}
	}
}

func TestSafePathValidation(t *testing.T) {
	// 测试安全路径验证
	tests := []struct {
		path       string
		shouldPass bool
	}{
		{"./test.txt", true},
		{"../test.txt", false}, // 上级目录应该失败
		{"/etc/passwd", false}, // 系统文件应该失败
		{"test.txt", true},
	}

	for _, tt := range tests {
		err := validateSafePath(tt.path)
		if tt.shouldPass && err != nil {
			t.Errorf("路径 %s 应该通过验证但失败: %v", tt.path, err)
		}
		if !tt.shouldPass && err == nil {
			t.Errorf("路径 %s 应该失败但通过了验证", tt.path)
		}
	}
}
