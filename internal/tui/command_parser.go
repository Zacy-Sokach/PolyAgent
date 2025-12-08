package tui

import (
	"fmt"
	"regexp"
	"strings"
)

// CommandType 命令类型
type CommandType int

const (
	CommandTypeUnknown CommandType = iota
	CommandTypeEdit
	CommandTypeTaskAdd
	CommandTypeTaskComplete
	CommandTypeTaskStart
	CommandTypeTaskCancel
	CommandTypeTaskRemove
	CommandTypeTaskClear
	CommandTypePlanUpdate
	CommandTypeInit
)

// Command 解析后的命令
type Command struct {
	Type        CommandType
	Raw         string
	Content     string
	TaskNumber  int
	Priority    string
	Description string
}

// CommandParser 命令解析器
type CommandParser struct {
	editPatterns         []*regexp.Regexp
	taskAddPatterns      []*regexp.Regexp
	taskCompletePatterns []*regexp.Regexp
	taskStartPatterns    []*regexp.Regexp
	taskCancelPatterns   []*regexp.Regexp
	taskRemovePatterns   []*regexp.Regexp
	taskClearPatterns    []*regexp.Regexp
	planUpdatePatterns   []*regexp.Regexp
	initPatterns         []*regexp.Regexp
}

// NewCommandParser 创建新的命令解析器
func NewCommandParser() *CommandParser {
	parser := &CommandParser{}
	parser.initializePatterns()
	return parser
}

// initializePatterns 初始化正则表达式模式
func (p *CommandParser) initializePatterns() {
	// 编辑命令模式
	p.editPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^EDIT\s+(.+)$`),
		regexp.MustCompile(`在文件\s+(.+?)\s+(插入|删除|替换)`),
		regexp.MustCompile(`(?i)edit\s+(.+)$`),
	}

	// 任务添加模式
	p.taskAddPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^TASK\s+ADD\s+(\S+)(?:\s+(\S+))?\s+(.+)$`),
		regexp.MustCompile(`添加任务\s*[:：]?\s*(.+?)(?:\s+优先级\s*[:：]?\s*(\S+))?$`),
		regexp.MustCompile(`(?i)add\s+task\s+(.+?)(?:\s+priority\s+(\S+))?$`),
	}

	// 任务完成模式
	p.taskCompletePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^TASK\s+COMPLETE\s+(\d+)$`),
		regexp.MustCompile(`完成任务\s*(\d+)`),
		regexp.MustCompile(`(?i)complete\s+task\s+(\d+)`),
	}

	// 任务开始模式
	p.taskStartPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^TASK\s+START\s+(\d+)$`),
		regexp.MustCompile(`开始任务\s*(\d+)`),
		regexp.MustCompile(`(?i)start\s+task\s+(\d+)`),
	}

	// 任务取消模式
	p.taskCancelPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^TASK\s+CANCEL\s+(\d+)$`),
		regexp.MustCompile(`取消任务\s*(\d+)`),
		regexp.MustCompile(`(?i)cancel\s+task\s+(\d+)`),
	}

	// 任务移除模式
	p.taskRemovePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^TASK\s+REMOVE\s+(\d+)$`),
		regexp.MustCompile(`移除任务\s*(\d+)`),
		regexp.MustCompile(`(?i)remove\s+task\s+(\d+)`),
	}

	// 任务清空模式
	p.taskClearPatterns = []*regexp.Regexp{
		regexp.MustCompile(`清空任务`),
		regexp.MustCompile(`重置任务`),
		regexp.MustCompile(`(?i)clear\s+tasks`),
		regexp.MustCompile(`(?i)reset\s+tasks`),
	}

	// 计划更新模式
	p.planUpdatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^PLAN\s+UPDATE\s+(.+)$`),
		regexp.MustCompile(`更新计划文档\s*[:：]?\s*(.+)`),
		regexp.MustCompile(`(?i)update\s+plan\s+(.+)`),
	}

	// init 命令模式（使用 /init 格式避免误触）
	p.initPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^/init$`),
		regexp.MustCompile(`^/init\s*$`),
	}
}

// Parse 解析命令字符串
func (p *CommandParser) Parse(input string) *Command {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// 检查编辑命令
	for _, pattern := range p.editPatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			return &Command{
				Type:    CommandTypeEdit,
				Raw:     input,
				Content: strings.TrimSpace(matches[1]),
			}
		}
	}

	// 检查任务添加命令
	for _, pattern := range p.taskAddPatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			cmd := &Command{
				Type: CommandTypeTaskAdd,
				Raw:  input,
			}

			if len(matches) >= 2 {
				cmd.Description = strings.TrimSpace(matches[1])
			}
			if len(matches) >= 3 && matches[2] != "" {
				cmd.Priority = strings.ToLower(strings.TrimSpace(matches[2]))
			} else {
				cmd.Priority = "medium"
			}

			return cmd
		}
	}

	// 检查任务完成命令
	for _, pattern := range p.taskCompletePatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			taskNum := 0
			fmt.Sscanf(matches[1], "%d", &taskNum)
			return &Command{
				Type:       CommandTypeTaskComplete,
				Raw:        input,
				TaskNumber: taskNum,
			}
		}
	}

	// 检查任务开始命令
	for _, pattern := range p.taskStartPatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			taskNum := 0
			fmt.Sscanf(matches[1], "%d", &taskNum)
			return &Command{
				Type:       CommandTypeTaskStart,
				Raw:        input,
				TaskNumber: taskNum,
			}
		}
	}

	// 检查任务取消命令
	for _, pattern := range p.taskCancelPatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			taskNum := 0
			fmt.Sscanf(matches[1], "%d", &taskNum)
			return &Command{
				Type:       CommandTypeTaskCancel,
				Raw:        input,
				TaskNumber: taskNum,
			}
		}
	}

	// 检查任务移除命令
	for _, pattern := range p.taskRemovePatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			taskNum := 0
			fmt.Sscanf(matches[1], "%d", &taskNum)
			return &Command{
				Type:       CommandTypeTaskRemove,
				Raw:        input,
				TaskNumber: taskNum,
			}
		}
	}

	// 检查任务清空命令
	for _, pattern := range p.taskClearPatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeTaskClear,
				Raw:  input,
			}
		}
	}

	// 检查计划更新命令
	for _, pattern := range p.planUpdatePatterns {
		if matches := pattern.FindStringSubmatch(input); matches != nil {
			return &Command{
				Type:    CommandTypePlanUpdate,
				Raw:     input,
				Content: strings.TrimSpace(matches[1]),
			}
		}
	}

	// 检查 init 命令
	for _, pattern := range p.initPatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeInit,
				Raw:  input,
			}
		}
	}

	return nil
}

// IsCommand 检查字符串是否为命令
func (p *CommandParser) IsCommand(input string) bool {
	return p.Parse(input) != nil
}

// FormatCommandType 格式化命令类型为字符串
func FormatCommandType(cmdType CommandType) string {
	switch cmdType {
	case CommandTypeEdit:
		return "EDIT"
	case CommandTypeTaskAdd:
		return "TASK_ADD"
	case CommandTypeTaskComplete:
		return "TASK_COMPLETE"
	case CommandTypeTaskStart:
		return "TASK_START"
	case CommandTypeTaskCancel:
		return "TASK_CANCEL"
	case CommandTypeTaskRemove:
		return "TASK_REMOVE"
	case CommandTypeTaskClear:
		return "TASK_CLEAR"
	case CommandTypePlanUpdate:
		return "PLAN_UPDATE"
	case CommandTypeInit:
		return "INIT"
	default:
		return "UNKNOWN"
	}
}
