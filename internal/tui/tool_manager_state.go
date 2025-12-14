package tui

import (
	"fmt"
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

// ToolManagerState 管理工具和任务状态
type ToolManagerState struct {
	editor           *utils.Editor
	tasks            []Task
	planDoc          PlanDoc
	currentTaskIndex int
	toolManager      interface{} // 避免循环引用
	commandParser    *CommandParser
}

// NewToolManagerState 创建新的工具管理器状态
func NewToolManagerState(toolManager interface{}, commandParser *CommandParser) *ToolManagerState {
	editor := utils.NewEditor()
	// 安全地初始化编辑器，捕获可能的panic
	var initError error
	func() {
		defer func() {
			if r := recover(); r != nil {
				initError = fmt.Errorf("编辑器初始化时发生错误: %v", r)
			}
		}()
		if err := editor.StartSession(); err != nil {
			initError = fmt.Errorf("初始化编辑会话失败: %w", err)
		}
	}()

	if initError != nil {
		// 可以选择记录错误或使用默认值
	}

	return &ToolManagerState{
		editor:           editor,
		tasks:            []Task{},
		planDoc:          PlanDoc{Version: 0, UpdatedAt: time.Now()},
		currentTaskIndex: -1,
		toolManager:      toolManager,
		commandParser:    commandParser,
	}
}

// GetEditor 获取编辑器
func (m *ToolManagerState) GetEditor() *utils.Editor {
	return m.editor
}

// GetTasks 获取任务列表
func (m *ToolManagerState) GetTasks() []Task {
	return m.tasks
}

// SetTasks 设置任务列表
func (m *ToolManagerState) SetTasks(tasks []Task) {
	m.tasks = tasks
}

// AddTask 添加任务
func (m *ToolManagerState) AddTask(task Task) {
	m.tasks = append(m.tasks, task)
}

// GetPlanDoc 获取计划文档
func (m *ToolManagerState) GetPlanDoc() PlanDoc {
	return m.planDoc
}

// SetPlanDoc 设置计划文档
func (m *ToolManagerState) SetPlanDoc(planDoc PlanDoc) {
	m.planDoc = planDoc
}

// GetCurrentTaskIndex 获取当前任务索引
func (m *ToolManagerState) GetCurrentTaskIndex() int {
	return m.currentTaskIndex
}

// SetCurrentTaskIndex 设置当前任务索引
func (m *ToolManagerState) SetCurrentTaskIndex(index int) {
	m.currentTaskIndex = index
}

// GetToolManager 获取工具管理器
func (m *ToolManagerState) GetToolManager() interface{} {
	return m.toolManager
}

// GetCommandParser 获取命令解析器
func (m *ToolManagerState) GetCommandParser() *CommandParser {
	return m.commandParser
}