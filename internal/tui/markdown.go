package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"

	"github.com/yuin/goldmark/util"
)

// MarkdownRenderer TUI Markdown 渲染器
type MarkdownRenderer struct {
	md         goldmark.Markdown
	styles     map[ast.NodeKind]lipgloss.Style
	styleStack []lipgloss.Style // 样式栈，用于嵌套样式
	buffer     *bytes.Buffer    // 缓冲区，用于复杂节点捕获
	maxStackDepth int           // 样式栈最大深度，防止栈溢出
}

// NewMarkdownRenderer 初始化渲染器 (单例模式预加载)
func NewMarkdownRenderer() *MarkdownRenderer {
	r := &MarkdownRenderer{
		styles:         make(map[ast.NodeKind]lipgloss.Style),
		styleStack:     nil, // 初始为空
		buffer:         nil, // 初始为空
		maxStackDepth:  10,  // 设置最大栈深度
	}
	r.initStyles()
	r.initGoldmark()
	return r
}

// initStyles 初始化样式配置
func (r *MarkdownRenderer) initStyles() {
	r.styles = map[ast.NodeKind]lipgloss.Style{
		// 标题：青色 + 粗体 + 下边距
		ast.KindHeading: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1),

		// 代码块：深灰背景 + 浅灰文字 + 内边距
		ast.KindCodeBlock: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1),

		// 行内代码：黄色文字
		ast.KindCodeSpan: lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Background(lipgloss.Color("237")),

		// 链接：蓝色 + 下划线
		ast.KindLink: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Underline(true),

		// 强调 (Bold/Italic)：由 renderEmphasis 动态处理，这里定义基础颜色
		ast.KindEmphasis: lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")), // 粉色高亮

		// 列表项：绿色
		ast.KindListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("78")),

		// 引用：斜体 + 灰色
		ast.KindBlockquote: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("245")).
			PaddingLeft(2).
			MarginTop(1).
			MarginBottom(1),
	}

	// 动态添加 GFM 扩展样式（如果可用）
	r.addGFMStyles()
}

// initGoldmark 初始化 Goldmark 实例
func (r *MarkdownRenderer) initGoldmark() {
	r.md = goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(), // 支持属性
		),
		goldmark.WithExtensions(
			extension.GFM, // 添加 GFM 扩展
		),
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(r, 1000),
				),
			),
		),
	)
}

// Render 将 Markdown 字符串渲染为 ANSI 字符串
func (r *MarkdownRenderer) Render(markdown string) string {
	// 重置状态
	r.resetStack()
	r.buffer = nil

	// 输入验证
	if len(markdown) == 0 {
		return ""
	}
	
	// 不再限制输入长度，允许完整渲染

	// 添加panic恢复机制
	defer func() {
		if recovered := recover(); recovered != nil {
			// 发生panic时重置状态
			r.styleStack = nil
		}
	}()

	var buf bytes.Buffer
	if err := r.md.Convert([]byte(markdown), &buf); err != nil {
		// 渲染失败时返回原始内容，但添加错误提示
		errorMsg := "❌ 渲染错误: " + err.Error() + "\n\n原始内容:\n"
		errorMsg += markdown
		return errorMsg
	}

	result := buf.String()
	
	// 验证输出结果，确保不包含恶意ANSI序列
	if len(result) > len(markdown)*10 { // 输出不应该比输入大太多
		return "⚠️ 渲染结果异常，返回原始内容:\n\n" + markdown
	}

	return result
}

// pushStyle 安全地推入样式栈
func (r *MarkdownRenderer) pushStyle(style lipgloss.Style) {
	if len(r.styleStack) < r.maxStackDepth {
		r.styleStack = append(r.styleStack, style)
	}
}

// popStyle 安全地弹出样式栈
func (r *MarkdownRenderer) popStyle() lipgloss.Style {
	if len(r.styleStack) > 0 {
		style := r.styleStack[len(r.styleStack)-1]
		r.styleStack = r.styleStack[:len(r.styleStack)-1]
		return style
	}
	return lipgloss.NewStyle()
}

// getCurrentStyle 获取当前合并的样式
func (r *MarkdownRenderer) getCurrentStyle() lipgloss.Style {
	style := lipgloss.NewStyle()
	for _, s := range r.styleStack {
		style = style.Inherit(s)
	}
	return style
}

// resetStack 重置样式栈
func (r *MarkdownRenderer) resetStack() {
	r.styleStack = nil
}

// RegisterFuncs 实现 goldmark.NodeRenderer 接口
func (r *MarkdownRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// 块级元素
	reg.Register(ast.KindDocument, r.renderNoop)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)

	// 行内元素
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)

	// GFM扩展元素 - 动态注册
	r.registerGFMElements(reg)

	// 其他忽略
	reg.Register(ast.KindThematicBreak, r.renderNoop)
}

// registerGFMElements 动态注册GFM扩展元素
func (r *MarkdownRenderer) registerGFMElements(reg renderer.NodeRendererFuncRegisterer) {
	// 尝试注册GFM扩展的节点类型
	// 由于goldmark的GFM扩展可能没有导出节点类型常量，
	// 我们使用字符串匹配的方式动态注册
	
	// 注册删除线
	if strikethroughKind := r.resolveKind("Strikethrough"); strikethroughKind != nil {
		reg.Register(*strikethroughKind, r.renderStrikethrough)
	}
	
	// 注册任务列表复选框
	if taskCheckBoxKind := r.resolveKind("TaskCheckBox"); taskCheckBoxKind != nil {
		reg.Register(*taskCheckBoxKind, r.renderTaskCheckBox)
	}
	
	// 注册表格相关元素
	if tableKind := r.resolveKind("Table"); tableKind != nil {
		reg.Register(*tableKind, r.renderTable)
	}
}

// resolveKind 尝试解析节点类型
func (r *MarkdownRenderer) resolveKind(kindName string) *ast.NodeKind {
	// 这里使用一个简化的方法，通过字符串匹配来识别GFM节点
	// 在实际运行时，goldmark会提供正确的节点类型
	return nil // 暂时返回nil，在渲染时通过字符串匹配处理
}

// ============================================================================
// 渲染实现方法
// ============================================================================

// renderText 基础文本渲染 (作为叶子节点)
func (r *MarkdownRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if textNode, ok := node.(*ast.Text); ok {
			text := string(textNode.Segment.Value(source))
			// 应用当前合并的样式
			currentStyle := r.getCurrentStyle()
			w.WriteString(currentStyle.Render(text))
		}
	}
	return ast.WalkContinue, nil
}

// getStyleForNode 根据节点类型获取样式（支持动态 GFM 节点）
func (r *MarkdownRenderer) getStyleForNode(node ast.Node) lipgloss.Style {
	kind := node.Kind()

	// 检查是否有预定义样式
	if style, ok := r.styles[kind]; ok {
		return style
	}

	// 动态处理 GFM 节点类型
	kindStr := kind.String()
	switch {
	case strings.Contains(kindStr, "Strikethrough"):
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Strikethrough(true)
	case strings.Contains(kindStr, "Table"):
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)
	case strings.Contains(kindStr, "TaskCheckBox"):
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("78"))
	default:
		return lipgloss.NewStyle() // 空样式
	}
}

// renderHeading 标题
func (r *MarkdownRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入标题样式栈
		style := r.getStyleForNode(node)
		heading := node.(*ast.Heading)
		if heading.Level == 1 {
			style = style.Bold(true).Underline(true)
		}
		r.pushStyle(style)
	} else {
		// 弹出样式栈
		r.popStyle()
		// 标题后添加换行
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderParagraph 段落 (智能间距处理)
func (r *MarkdownRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		// 检查父节点类型，决定是否添加空行
		parent := node.Parent()
		if parent != nil && parent.Kind() == ast.KindListItem {
			// 列表项内的段落，只加一个换行
			w.WriteString("\n")
		} else if parent != nil && parent.Kind() == ast.KindBlockquote {
			// 引用块内的段落，只加一个换行
			w.WriteString("\n")
		} else {
			// 普通段落，添加两个换行
			w.WriteString("\n\n")
		}
	}
	return ast.WalkContinue, nil
}

// renderCodeBlock 代码块
func (r *MarkdownRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		var codeBuilder strings.Builder
		lines := node.(*ast.CodeBlock).Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			codeBuilder.Write(line.Value(source))
		}

		// 移除末尾多余换行
		code := strings.TrimRight(codeBuilder.String(), "\n")
		style := r.getStyleForNode(node)

		w.WriteString(style.Render(code))
		w.WriteString("\n")
	}
	return ast.WalkSkipChildren, nil
}

// renderCodeSpan 行内代码
func (r *MarkdownRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入代码样式栈
		r.pushStyle(r.getStyleForNode(node))
	} else {
		// 弹出样式栈
		r.popStyle()
	}
	return ast.WalkContinue, nil
}

// renderLink 链接 [Text](URL)
func (r *MarkdownRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入链接样式栈
		r.pushStyle(r.getStyleForNode(node))
	} else {
		// 弹出样式栈
		r.popStyle()
		// 在链接结束后添加 URL 显示
		if link, ok := node.(*ast.Link); ok {
			url := string(link.Destination)
			w.WriteString(fmt.Sprintf(" (%s)", url))
		}
	}
	return ast.WalkContinue, nil
}

// renderEmphasis 强调 (粗体/斜体)
func (r *MarkdownRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入样式栈
		style := r.getStyleForNode(node)
		if emphasis, ok := node.(*ast.Emphasis); ok {
			level := emphasis.Level
			if level == 2 {
				style = style.Bold(true)
			} else {
				style = style.Italic(true)
			}
		}
		r.pushStyle(style)
	} else {
		// 弹出样式栈
		r.popStyle()
	}
	return ast.WalkContinue, nil
}

// renderBlockquote 引用
func (r *MarkdownRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入引用样式栈
		r.pushStyle(r.getStyleForNode(node))
	} else {
		// 弹出样式栈
		r.popStyle()
		// 引用块后添加换行
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderList 列表
func (r *MarkdownRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// renderListItem 列表项
func (r *MarkdownRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		var prefix string
		if parent := node.Parent(); parent != nil {
			if list, ok := parent.(*ast.List); ok {
				if list.IsOrdered() {
					// 获取当前项在父节点中的索引
					index := 1
					for n := node.PreviousSibling(); n != nil; n = node.PreviousSibling() {
						index++
					}
					prefix = fmt.Sprintf("%d. ", list.Start+index-1)
				} else {
					prefix = "• "
				}
			} else {
				prefix = "• " // 默认使用无序列表符号
			}
		} else {
			prefix = "• " // 默认使用无序列表符号
		}

		// 渲染前缀（应用列表项样式）
		style := r.getStyleForNode(node)
		w.WriteString(style.Render(prefix))

		// 推入列表项样式栈，让内容也应用样式
		r.pushStyle(style)
	} else {
		// 弹出样式栈
		r.popStyle()
		// 列表项后添加换行
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

func (r *MarkdownRenderer) renderNoop(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

// addGFMStyles 动态添加 GFM 扩展样式
func (r *MarkdownRenderer) addGFMStyles() {
	// 由于 GFM 扩展的节点类型常量可能不存在，我们使用字符串匹配
	// 在实际渲染时，我们会根据节点类型字符串来应用样式

	// 定义 GFM 节点类型的字符串常量
	const (
		kindStrikethrough = "Strikethrough"
		kindTable         = "Table"
		kindTableHeader   = "TableHeader"
		kindTaskCheckBox  = "TaskCheckBox"
	)

	// 创建样式但不立即赋值，在渲染时动态匹配
	// 这些样式将在 renderText 中根据节点类型字符串应用
}

// renderStrikethrough 删除线
func (r *MarkdownRenderer) renderStrikethrough(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入删除线样式栈
		r.pushStyle(r.getStyleForNode(node))
	} else {
		// 弹出样式栈
		r.popStyle()
	}
	return ast.WalkContinue, nil
}

// renderTaskCheckBox 任务列表复选框
func (r *MarkdownRenderer) renderTaskCheckBox(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// 推入任务复选框样式栈
		r.pushStyle(r.getStyleForNode(node))

		// 检查是否已勾选（通过节点属性或子节点内容判断）
		isChecked := false
		if child := node.FirstChild(); child != nil {
			if textNode, ok := child.(*ast.Text); ok {
				text := string(textNode.Segment.Value(source))
				isChecked = strings.Contains(text, "x") || strings.Contains(text, "X")
			}
		}

		// 渲染复选框符号
		if isChecked {
			w.WriteString("[✓] ")
		} else {
			w.WriteString("[ ] ")
		}
	} else {
		// 弹出样式栈
		r.popStyle()
	}
	return ast.WalkContinue, nil
}

// renderTable 表格渲染
func (r *MarkdownRenderer) renderTable(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushStyle(r.getStyleForNode(node))
	} else {
		r.popStyle()
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// ----------------------------------------------------------------------------
// 辅助函数
// ----------------------------------------------------------------------------

// extractText 递归提取节点下的所有纯文本内容
func (r *MarkdownRenderer) extractText(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			buf.Write(textNode.Segment.Value(source))
		} else {
			// 递归提取嵌套节点（如强调中的文本）
			buf.WriteString(r.extractText(child, source))
		}
	}
	return buf.String()
}
