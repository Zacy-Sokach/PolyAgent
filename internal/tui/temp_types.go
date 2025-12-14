package tui

import (
	"time"
)

// 临时类型定义，用于编译验证

// Message 表示对话中的消息
type Message struct {
	Role    string
	Content string
}

// Task 表示任务管理中的任务
type Task struct {
	ID          string
	Description string
	Status      string // "pending", "in_progress", "completed", "cancelled"
	Priority    string // "high", "medium", "low"
}

// PlanDoc 表示计划文档
type PlanDoc struct {
	Content   string
	Version   int
	UpdatedAt time.Time
}

// ToolManager 和 CommandParser 已在其他文件中定义