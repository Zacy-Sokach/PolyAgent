package tui

import (
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

// Message types for tea.Model
type ResponseMsg struct {
	Content string
}

type StreamChunkMsg struct {
	Chunk     string
	Reasoning string
}

type ToolCallMsg struct {
	ToolCalls []api.ToolCall
}

type ToolResultMsg struct {
	ResultMessages []api.Message
	DisplayContent string
}

// Additional message types for the model
type StartStreamMsg struct {
	Input string
}

type CheckStreamMsg struct{}

type StreamErrorMsg struct {
	Error error
}

type SaveHistoryMsg struct{}

type ClearHistoryMsg struct{}

type LoadHistoryMsg struct {
	Messages []Message
}

type UpdatePlanMsg struct {
	Content string
}

type TaskUpdateMsg struct {
	TaskID string
	Status string
}

type EditorInitMsg struct {
	Editor *utils.Editor
}

type EditorUpdateMsg struct {
	Content string
}

type ExportMarkdownMsg struct {
	FilePath string
}

type ExportSuccessMsg struct {
	FilePath string
}

type ExportErrorMsg struct {
	Error error
}

// Tick message for auto-save or other periodic tasks
type TickMsg struct {
	Time time.Time
}
