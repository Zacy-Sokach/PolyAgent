package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

// EditCommand AI编辑指令
type EditCommand struct {
	Type     string // "insert", "delete", "replace", "create"
	FilePath string
	Offset   int
	Length   int
	Content  string
}

// EditorParser 编辑指令解析器
type EditorParser struct {
	editor *utils.Editor
}

// NewEditorParser 创建新的编辑指令解析器
func NewEditorParser(editor *utils.Editor) *EditorParser {
	return &EditorParser{
		editor: editor,
	}
}

// ParseAndExecute 解析并执行编辑指令
func (p *EditorParser) ParseAndExecute(commandText string) (string, error) {
	// 尝试解析为结构化指令
	if cmd, ok := p.parseStructuredCommand(commandText); ok {
		return p.executeCommand(cmd)
	}

	// 尝试解析为自然语言指令
	if cmd, ok := p.parseNaturalLanguage(commandText); ok {
		return p.executeCommand(cmd)
	}

	return "", fmt.Errorf("无法解析编辑指令: %s", commandText)
}

// parseStructuredCommand 解析结构化指令
// 格式: EDIT <type> <file> <offset> <length> <content>
func (p *EditorParser) parseStructuredCommand(text string) (*EditCommand, bool) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToUpper(text), "EDIT ") {
		return nil, false
	}

	parts := strings.SplitN(text, " ", 6)
	if len(parts) < 4 {
		return nil, false
	}

	cmd := &EditCommand{
		Type: strings.ToLower(parts[1]),
	}

	switch cmd.Type {
	case "insert", "delete", "replace":
		if len(parts) < 5 {
			return nil, false
		}
		cmd.FilePath = parts[2]

		offset, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, false
		}
		cmd.Offset = offset

		switch cmd.Type {
		case "insert":
			if len(parts) >= 6 {
				cmd.Content = parts[5]
			}
		case "delete":
			length, err := strconv.Atoi(parts[4])
			if err != nil {
				return nil, false
			}
			cmd.Length = length
		case "replace":
			length, err := strconv.Atoi(parts[4])
			if err != nil {
				return nil, false
			}
			cmd.Length = length
			if len(parts) >= 6 {
				cmd.Content = parts[5]
			}
		}

	case "create":
		if len(parts) < 4 {
			return nil, false
		}
		cmd.FilePath = parts[2]
		if len(parts) >= 4 {
			cmd.Content = strings.Join(parts[3:], " ")
		}

	default:
		return nil, false
	}

	return cmd, true
}

// parseNaturalLanguage 解析自然语言指令
func (p *EditorParser) parseNaturalLanguage(text string) (*EditCommand, bool) {
	text = strings.ToLower(text)

	// 匹配插入模式
	if matches := regexp.MustCompile(`在文件\s+([^\s]+)\s+的第\s+(\d+)\s+行(?:到第\s+(\d+)\s+行)?\s*(?:插入|添加|加上)\s*:\s*(.+)`).FindStringSubmatch(text); matches != nil {
		cmd := &EditCommand{
			Type:     "insert",
			FilePath: matches[1],
		}

		// 获取文件内容以计算偏移量
		content, err := p.editor.GetFileContent(matches[1])
		if err != nil {
			return nil, false
		}

		lineNum, _ := strconv.Atoi(matches[2])
		cmd.Offset = p.lineToOffset(content, lineNum)
		cmd.Content = matches[4]

		return cmd, true
	}

	// 匹配删除模式
	if matches := regexp.MustCompile(`删除文件\s+([^\s]+)\s+的第\s+(\d+)\s+行(?:到第\s+(\d+)\s+行)?`).FindStringSubmatch(text); matches != nil {
		cmd := &EditCommand{
			Type:     "delete",
			FilePath: matches[1],
		}

		content, err := p.editor.GetFileContent(matches[1])
		if err != nil {
			return nil, false
		}

		startLine, _ := strconv.Atoi(matches[2])
		cmd.Offset = p.lineToOffset(content, startLine)

		if matches[3] != "" {
			endLine, _ := strconv.Atoi(matches[3])
			endOffset := p.lineToOffset(content, endLine+1) // +1 到下一行开始
			cmd.Length = endOffset - cmd.Offset
		} else {
			// 删除单行
			lines := strings.Split(content, "\n")
			if startLine-1 < len(lines) {
				lineContent := lines[startLine-1]
				cmd.Length = len(lineContent)
				if startLine < len(lines) {
					cmd.Length += 1 // 包括换行符
				}
			}
		}

		return cmd, true
	}

	// 匹配替换模式
	if matches := regexp.MustCompile(`将文件\s+([^\s]+)\s+的第\s+(\d+)\s+行(?:到第\s+(\d+)\s+行)?\s*替换为\s*:\s*(.+)`).FindStringSubmatch(text); matches != nil {
		cmd := &EditCommand{
			Type:     "replace",
			FilePath: matches[1],
			Content:  matches[4],
		}

		content, err := p.editor.GetFileContent(matches[1])
		if err != nil {
			return nil, false
		}

		startLine, _ := strconv.Atoi(matches[2])
		cmd.Offset = p.lineToOffset(content, startLine)

		if matches[3] != "" {
			endLine, _ := strconv.Atoi(matches[3])
			endOffset := p.lineToOffset(content, endLine+1)
			cmd.Length = endOffset - cmd.Offset
		} else {
			// 替换单行
			lines := strings.Split(content, "\n")
			if startLine-1 < len(lines) {
				lineContent := lines[startLine-1]
				cmd.Length = len(lineContent)
				if startLine < len(lines) {
					cmd.Length += 1 // 包括换行符
				}
			}
		}

		return cmd, true
	}

	return nil, false
}

// executeCommand 执行编辑命令
func (p *EditorParser) executeCommand(cmd *EditCommand) (string, error) {
	switch cmd.Type {
	case "insert":
		if err := p.editor.InsertText(cmd.FilePath, cmd.Offset, cmd.Content); err != nil {
			return "", err
		}
		return fmt.Sprintf("已在文件 %s 的偏移量 %d 处插入内容", cmd.FilePath, cmd.Offset), nil

	case "delete":
		if err := p.editor.DeleteText(cmd.FilePath, cmd.Offset, cmd.Length); err != nil {
			return "", err
		}
		return fmt.Sprintf("已删除文件 %s 中从偏移量 %d 开始的 %d 个字符", cmd.FilePath, cmd.Offset, cmd.Length), nil

	case "replace":
		if err := p.editor.ReplaceText(cmd.FilePath, cmd.Offset, cmd.Length, cmd.Content); err != nil {
			return "", err
		}
		return fmt.Sprintf("已替换文件 %s 中从偏移量 %d 开始的 %d 个字符", cmd.FilePath, cmd.Offset, cmd.Length), nil

	case "create":
		// 创建新文件需要先写入内容到磁盘，然后重新加载到编辑器
		content := cmd.Content
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}

		// 使用现有的文件操作工具
		if err := mcp.CreateNewFile(cmd.FilePath, content); err != nil {
			return "", err
		}

		// 重新加载文件到编辑器
		if err := p.editor.LoadFile(cmd.FilePath); err != nil {
			return "", err
		}

		return fmt.Sprintf("已创建文件 %s", cmd.FilePath), nil

	default:
		return "", fmt.Errorf("未知命令类型: %s", cmd.Type)
	}
}

// lineToOffset 将行号转换为字符偏移量
func (p *EditorParser) lineToOffset(content string, lineNum int) int {
	if lineNum <= 1 {
		return 0
	}

	lines := strings.Split(content, "\n")
	offset := 0

	for i := 0; i < lineNum-1 && i < len(lines); i++ {
		offset += len(lines[i]) + 1 // +1 为换行符
	}

	return offset
}

// LoadFile 加载文件到编辑器（辅助方法）
func (p *EditorParser) LoadFile(filePath string) error {
	// 这个方法需要在 Editor 中添加
	// 暂时使用反射或其他方式调用
	return nil
}
