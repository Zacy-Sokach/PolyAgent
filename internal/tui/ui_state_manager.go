package tui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

// UIStateManager 管理UI组件的状态
type UIStateManager struct {
	viewport viewport.Model
	textarea textarea.Model
	ready    bool
}

// NewUIStateManager 创建新的UI状态管理器
func NewUIStateManager() *UIStateManager {
	ta := textarea.New()
	ta.Placeholder = "输入你的问题..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(80, 20)
	vp.SetContent("欢迎使用 PolyAgent - 类似 Claude Code 的 Vibe Coding 工具\n\n")

	return &UIStateManager{
		viewport: vp,
		textarea: ta,
		ready:    false,
	}
}

// GetViewport 获取viewport组件
func (m *UIStateManager) GetViewport() viewport.Model {
	return m.viewport
}

// SetViewport 设置viewport组件
func (m *UIStateManager) SetViewport(vp viewport.Model) {
	m.viewport = vp
}

// GetTextarea 获取textarea组件
func (m *UIStateManager) GetTextarea() textarea.Model {
	return m.textarea
}

// SetTextarea 设置textarea组件
func (m *UIStateManager) SetTextarea(ta textarea.Model) {
	m.textarea = ta
}

// IsReady 检查是否已准备就绪
func (m *UIStateManager) IsReady() bool {
	return m.ready
}

// SetReady 设置准备状态
func (m *UIStateManager) SetReady(ready bool) {
	m.ready = ready
}

// UpdateViewportSize 更新viewport尺寸
func (m *UIStateManager) UpdateViewportSize(width, height int) {
	if !m.ready {
		m.viewport = viewport.New(width, height-4)
		m.viewport.YPosition = 0
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = height - 4
	}
	m.textarea.SetWidth(width)
}