package tui

import (
	"sync"
)

// 全局Markdown渲染器单例
var (
	globalMarkdownRenderer *CGOMarkdownRenderer
	rendererOnce           sync.Once
)

// GetMarkdownRenderer 获取Markdown渲染器单例
func GetMarkdownRenderer() *CGOMarkdownRenderer {
	rendererOnce.Do(func() {
		globalMarkdownRenderer = NewCGOMarkdownRenderer()
	})
	return globalMarkdownRenderer
}

// MarkdownRenderer TUI Markdown 渲染器接口
// 为了保持向后兼容，保留这个类型别名
type MarkdownRenderer = CGOMarkdownRenderer

// NewMarkdownRenderer 创建新的 Markdown 渲染器
// 为了保持向后兼容，保留这个函数
func NewMarkdownRenderer() *CGOMarkdownRenderer {
	return NewCGOMarkdownRenderer()
}