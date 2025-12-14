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
	CommandTypeClear
	CommandTypeInit
	CommandTypeCheckUpdate
	CommandTypeUpdate
	CommandTypeCoTEnable
	CommandTypeCoTDisable
	CommandTypeCoTToggle
	CommandTypeCoTHistory
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
	clearPatterns        []*regexp.Regexp
	initPatterns         []*regexp.Regexp
	checkUpdatePatterns  []*regexp.Regexp
	updatePatterns       []*regexp.Regexp
	cotEnablePatterns    []*regexp.Regexp
	cotDisablePatterns   []*regexp.Regexp
	cotTogglePatterns    []*regexp.Regexp
	cotHistoryPatterns   []*regexp.Regexp
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

	// clear 命令模式（必须使用 /clear 格式避免误触）
	p.clearPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^/clear$`),
		regexp.MustCompile(`^/clear\s*$`),
	}

	// init 命令模式（使用 /init 格式避免误触）
	p.initPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^/init$`),
		regexp.MustCompile(`^/init\s*$`),
	}

	// 检查更新命令模式
	p.checkUpdatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^check\s+update$`),
		regexp.MustCompile(`(?i)^检查更新$`),
		regexp.MustCompile(`^/check-update$`),
	}

	// 更新命令模式
	p.updatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^update$`),
		regexp.MustCompile(`(?i)^更新$`),
		regexp.MustCompile(`^/update$`),
	}

	// CoT启用命令模式
	p.cotEnablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^cot\s+enable$`),
		regexp.MustCompile(`(?i)^启用思考$`),
		regexp.MustCompile(`^/cot-enable$`),
	}

	// CoT禁用命令模式
	p.cotDisablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^cot\s+disable$`),
		regexp.MustCompile(`(?i)^禁用思考$`),
		regexp.MustCompile(`^/cot-disable$`),
	}

	// CoT切换命令模式
	p.cotTogglePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^cot\s+toggle$`),
		regexp.MustCompile(`(?i)^切换思考显示$`),
		regexp.MustCompile(`^/cot-toggle$`),
	}

	// CoT历史命令模式
	p.cotHistoryPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^cot\s+history$`),
		regexp.MustCompile(`(?i)^思考历史$`),
		regexp.MustCompile(`^/cot-history$`),
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

	// 检查 clear 命令
	for _, pattern := range p.clearPatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeClear,
				Raw:  input,
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

	// 检查更新命令
	for _, pattern := range p.checkUpdatePatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeCheckUpdate,
				Raw:  input,
			}
		}
	}

	// 检查更新命令
	for _, pattern := range p.updatePatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeUpdate,
				Raw:  input,
			}
		}
	}

	// 检查CoT启用命令
	for _, pattern := range p.cotEnablePatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeCoTEnable,
				Raw:  input,
			}
		}
	}

	// 检查CoT禁用命令
	for _, pattern := range p.cotDisablePatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeCoTDisable,
				Raw:  input,
			}
		}
	}

	// 检查CoT切换命令
	for _, pattern := range p.cotTogglePatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeCoTToggle,
				Raw:  input,
			}
		}
	}

	// 检查CoT历史命令
	for _, pattern := range p.cotHistoryPatterns {
		if pattern.MatchString(input) {
			return &Command{
				Type: CommandTypeCoTHistory,
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
	case CommandTypeClear:
		return "CLEAR"
	case CommandTypeInit:
		return "INIT"
	case CommandTypeCheckUpdate:
		return "CHECK_UPDATE"
	case CommandTypeUpdate:
		return "UPDATE"
	case CommandTypeCoTEnable:
		return "COT_ENABLE"
	case CommandTypeCoTDisable:
		return "COT_DISABLE"
	case CommandTypeCoTToggle:
		return "COT_TOGGLE"
	case CommandTypeCoTHistory:
		return "COT_HISTORY"
	default:
		return "UNKNOWN"
	}
}
