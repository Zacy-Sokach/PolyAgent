package tui

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L${SRCDIR}/rust_markdown/target/release -lrust_markdown -ldl -lm -static
#include <stdlib.h>
#include "c_markdown.h"
*/
import "C"
import (
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
)

// CGOMarkdownRenderer 使用 Rust pulldown-cmark 和 CGO 的 Markdown 渲染器
type CGOMarkdownRenderer struct {
	parser *C.MarkdownParser
	styles map[string]lipgloss.Style
	mu     sync.Mutex // 保证线程安全
}

// NewCGOMarkdownRenderer 创建新的 CGO Markdown 渲染器
func NewCGOMarkdownRenderer() *CGOMarkdownRenderer {
	parser := C.markdown_parser_new()
	if parser == nil {
		panic("Failed to create markdown parser")
	}

	r := &CGOMarkdownRenderer{
		parser: parser,
		styles: make(map[string]lipgloss.Style),
	}

	r.initStyles()
	r.setParserOptions()
	return r
}

// Free 释放 CGO 资源
func (r *CGOMarkdownRenderer) Free() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.parser != nil {
		C.markdown_parser_free(r.parser)
		r.parser = nil
	}
}

// initStyles 初始化样式配置
func (r *CGOMarkdownRenderer) initStyles() {
	// 这里保留样式定义以备将来扩展，但实际渲染由 Rust 库处理
	r.styles = make(map[string]lipgloss.Style)
}

// setParserOptions 设置解析器选项
func (r *CGOMarkdownRenderer) setParserOptions() {
	// 设置 GFM 选项
	C.markdown_set_gfm_enabled(r.parser, C.bool(true))
	C.markdown_set_table_enabled(r.parser, C.bool(true))
	C.markdown_set_strikethrough_enabled(r.parser, C.bool(true))
	C.markdown_set_tasklist_enabled(r.parser, C.bool(true))

	// 设置颜色 - 使用 defer 确保释放内存
	cHeadingColor := C.CString("86")
	C.markdown_set_heading_color(r.parser, cHeadingColor)
	C.free(unsafe.Pointer(cHeadingColor))

	cCodeColor := C.CString("252")
	C.markdown_set_code_color(r.parser, cCodeColor)
	C.free(unsafe.Pointer(cCodeColor))

	cLinkColor := C.CString("39")
	C.markdown_set_link_color(r.parser, cLinkColor)
	C.free(unsafe.Pointer(cLinkColor))

	cTextColor := C.CString("255")
	C.markdown_set_text_color(r.parser, cTextColor)
	C.free(unsafe.Pointer(cTextColor))
}

// Render 渲染 Markdown 文本为 ANSI 格式
func (r *CGOMarkdownRenderer) Render(markdown string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if markdown == "" {
		return ""
	}

	// 将 Go 字符串转换为 C 字符串
	cMarkdown := C.CString(markdown)
	defer C.free(unsafe.Pointer(cMarkdown))

	// 调用 Rust 库解析并渲染
	cResult := C.markdown_parse_to_ansi(r.parser, cMarkdown)
	if cResult == nil {
		// 使用无锁版本检查错误，避免死锁
		if r.hasErrorLocked() {
			return fmt.Sprintf("❌ 渲染错误: %s", r.getErrorLocked())
		}
		return ""
	}

	// 转换结果并释放内存
	result := C.GoString(cResult)
	C.markdown_free_string(cResult)

	// 后处理：只清理连续的三个以上换行
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")

	return result
}

// hasErrorLocked 检查是否有错误（内部无锁版本，调用者必须持有锁）
func (r *CGOMarkdownRenderer) hasErrorLocked() bool {
	return bool(C.markdown_has_error(r.parser))
}

// getErrorLocked 获取错误信息（内部无锁版本，调用者必须持有锁）
func (r *CGOMarkdownRenderer) getErrorLocked() string {
	cErr := C.markdown_get_error(r.parser)
	if cErr == nil {
		return ""
	}
	return C.GoString(cErr)
}

// HasError 检查是否有错误
func (r *CGOMarkdownRenderer) HasError() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hasErrorLocked()
}

// GetError 获取错误信息
func (r *CGOMarkdownRenderer) GetError() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getErrorLocked()
}

// SetGFMEnabled 设置 GFM 支持
func (r *CGOMarkdownRenderer) SetGFMEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	C.markdown_set_gfm_enabled(r.parser, C.bool(enabled))
}

// SetTableEnabled 设置表格支持
func (r *CGOMarkdownRenderer) SetTableEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	C.markdown_set_table_enabled(r.parser, C.bool(enabled))
}

// SetStrikethroughEnabled 设置删除线支持
func (r *CGOMarkdownRenderer) SetStrikethroughEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	C.markdown_set_strikethrough_enabled(r.parser, C.bool(enabled))
}

// SetTasklistEnabled 设置任务列表支持
func (r *CGOMarkdownRenderer) SetTasklistEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	C.markdown_set_tasklist_enabled(r.parser, C.bool(enabled))
}

// SetHeadingColor 设置标题颜色
func (r *CGOMarkdownRenderer) SetHeadingColor(color string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	C.markdown_set_heading_color(r.parser, cColor)
}

// SetCodeColor 设置代码颜色
func (r *CGOMarkdownRenderer) SetCodeColor(color string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	C.markdown_set_code_color(r.parser, cColor)
}

// SetLinkColor 设置链接颜色
func (r *CGOMarkdownRenderer) SetLinkColor(color string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	C.markdown_set_link_color(r.parser, cColor)
}

// SetTextColor 设置文本颜色
func (r *CGOMarkdownRenderer) SetTextColor(color string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	C.markdown_set_text_color(r.parser, cColor)
}

// truncateText 截断文本到指定宽度
func truncateText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	var result strings.Builder
	currentWidth := 0

	for _, r := range text {
		rw := runeWidth(r)
		if currentWidth+rw > width {
			break
		}
		result.WriteRune(r)
		currentWidth += rw
	}

	if result.Len() < len(text) {
		// 需要截断，添加省略号
		for currentWidth+3 > width && result.Len() > 0 {
			// 移除最后一个字符
			runes := []rune(result.String())
			if len(runes) > 0 {
				lastRune := runes[len(runes)-1]
				currentWidth -= runeWidth(lastRune)
				result.Reset()
				result.WriteString(string(runes[:len(runes)-1]))
			}
		}
		result.WriteString("...")
	}

	return result.String()
}

// runeWidth 计算 rune 的显示宽度
func runeWidth(r rune) int {
	if r < 32 || (r >= 0x7f && r <= 0x9f) {
		return 0
	}
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
		return 2
	}
	if r == utf8.RuneError {
		return 1
	}
	return 1
}
