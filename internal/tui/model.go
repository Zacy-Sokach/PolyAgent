package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
	"github.com/Zacy-Sokach/PolyAgent/internal/update"
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Version 是当前的 PolyAgent 版本，由 main 包设置
var Version string

// Message types for Bubble Tea
type CheckStreamMsg struct{}

type StreamChunkMsg struct {
	Chunk     string
	Reasoning string
}

type ResponseMsg struct {
	Content string
}

type ToolCallMsg struct {
	ToolCalls []api.ToolCall
}

type ToolResultMsg struct {
	ResultMessages []api.Message
	DisplayContent string
}

type StreamErrorMsg struct {
	Error error
}

type Message struct {
	Role    string
	Content string
}

type Task struct {
	ID          string
	Description string
	Status      string // "pending", "in_progress", "completed", "cancelled"
	Priority    string // "high", "medium", "low"
}

type PlanDoc struct {
	Content   string
	Version   int
	UpdatedAt time.Time
}

// ToolManager wraps MCP ToolRegistry for TUI usage
type ToolManager struct {
	registry *mcp.ToolRegistry
}

// NewToolManager creates a new ToolManager with default tools
func NewToolManager() *ToolManager {
	return &ToolManager{
		registry: mcp.DefaultToolRegistry(nil),
	}
}

// NewToolManagerWithRegistry creates a ToolManager with custom registry
func NewToolManagerWithRegistry(registry *mcp.ToolRegistry) *ToolManager {
	return &ToolManager{
		registry: registry,
	}
}

// GetToolsForAPI returns tools in API format
func (tm *ToolManager) GetToolsForAPI() []api.Tool {
	mcpTools := tm.registry.ListTools()
	tools := make([]api.Tool, len(mcpTools))
	
	for i, t := range mcpTools {
		tools[i] = api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		}
	}
	
	return tools
}

// HandleToolCalls executes tool calls and returns API messages
func (tm *ToolManager) HandleToolCalls(toolCalls []api.ToolCall) ([]api.Message, error) {
	var messages []api.Message
	
	for _, call := range toolCalls {
		// Convert json.RawMessage to map[string]interface{}
		var args map[string]interface{}
		if err := json.Unmarshal(call.Function.Arguments, &args); err != nil {
			// If unmarshaling fails, try to use as string
			args = map[string]interface{}{
				"input": string(call.Function.Arguments),
			}
		}
		
		// Convert to MCP request
		mcpRequest := mcp.CallToolRequest{
			Name:      call.Function.Name,
			Arguments: args,
		}
		
		// Execute via MCP registry
		result, err := tm.registry.HandleCallTool(mcpRequest)
		if err != nil {
			return nil, err
		}
		
		// Convert to API message
		if len(result.Content) > 0 {
			content := result.Content[0].Text
			messages = append(messages, api.ToolResultMessage(call.ID, content))
		}
	}
	
	return messages, nil
}

// FormatToolCallForDisplay formats tool call for UI display
func (tm *ToolManager) FormatToolCallForDisplay(call api.ToolCall) string {
	return fmt.Sprintf("🔧 调用工具: %s\n参数: %v", call.Function.Name, call.Function.Arguments)
}

type Model struct {
	viewport         viewport.Model
	textarea         textarea.Model
	messages         []Message
	ready            bool
	apiKey           string
	thinking         bool
	currentResp      string
	currentThink     string
	streamCh         <-chan string
	reasoningCh      <-chan string
	toolCallCh       <-chan []api.ToolCall
	streamErrCh      <-chan error
	editor           *utils.Editor
	tasks            []Task
	planDoc          PlanDoc
	currentTaskIndex int
	pendingToolCalls []api.ToolCall
	toolManager      *ToolManager
	apiMessages      []api.Message
	commandParser    *CommandParser
	maxMessages      int // 最大消息数量限制
	renderedLines    []string // 缓存已渲染的行，避免重复渲染
	lastRenderedHash uint64   // 上次渲染的内容哈希，用于检测变化
	ctx              context.Context // 用于取消操作的context
	cancel           context.CancelFunc // 取消函数
}

func InitialModel(apiKey string, toolManager *ToolManager) Model {
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

	editor := utils.NewEditor()
	// 安全地初始化编辑器，捕获可能的panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				vp.SetContent(fmt.Sprintf("编辑器初始化时发生错误: %v\n\n", r))
			}
		}()
		if err := editor.StartSession(); err != nil {
			vp.SetContent(fmt.Sprintf("初始化编辑会话失败: %v\n\n", err))
		}
	}()

	if toolManager == nil {
		toolManager = NewToolManager()
	}
	commandParser := NewCommandParser()

	// 创建context用于取消操作
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		textarea:         ta,
		viewport:         vp,
		messages:         []Message{},
		apiKey:           apiKey,
		editor:           editor,
		tasks:            []Task{},
		planDoc:          PlanDoc{Version: 0, UpdatedAt: time.Now()},
		currentTaskIndex: -1,
		toolManager:      toolManager,
		commandParser:    commandParser,
		maxMessages:      50,  // 限制最多显示50条消息
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.saveHistory()
			if m.editor != nil {
				m.editor.EndSession()
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.thinking {
				input := m.textarea.Value()
				if strings.TrimSpace(input) != "" {
					// 检查是否是命令
					if cmd := m.commandParser.Parse(input); cmd != nil {
						m.textarea.Reset()
						return m, m.handleCommand(cmd)
					}

					// 不是命令，发送给AI
					m.messages = append(m.messages, Message{Role: "user", Content: input})
					m.textarea.Reset()
					m.thinking = true
					m.currentResp = ""
					return m, tea.Batch(
						m.updateViewport(),
						m.startStream(input),
					)
				}
			}
		case tea.KeyCtrlS:
			if m.editor != nil {
				return m, m.saveChangesToDisk()
			}
		case tea.KeyEsc:
			if m.thinking {
				m.thinking = false
				// 取消正在进行的操作
				if m.cancel != nil {
					m.cancel()
				}
				// 重新创建context以便下次使用
				m.ctx, m.cancel = context.WithCancel(context.Background())
			}
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}
		m.textarea.SetWidth(msg.Width)

	case CheckStreamMsg:
		// 流结束了，更新历史消息缓存
		if len(m.pendingToolCalls) > 0 {
			// 如果有挂起的工具调用，不要停止思考，执行工具
			return m, m.executePendingTools()
		}

		m.thinking = false
		// 将累积的响应保存到消息历史中
		if m.currentResp != "" {
			m.messages = append(m.messages, Message{Role: "assistant", Content: m.currentResp})
			// 同时也保存到API历史
			m.apiMessages = append(m.apiMessages, api.TextMessage("assistant", m.currentResp))

			// 更新渲染缓存
			m.updateRenderedLinesCache()

			m.currentResp = ""
			m.currentThink = ""
			return m, m.updateViewport()
		}
		return m, nil

	case ResponseMsg:
		m.thinking = false
		m.messages = append(m.messages, Message{Role: "assistant", Content: msg.Content})
		m.currentThink = ""
		m.currentResp = ""
		return m, m.updateViewport()

	case StreamChunkMsg:
		if msg.Reasoning != "" {
			m.currentThink += msg.Reasoning
		} else {
			m.currentResp += msg.Chunk
		}
		
		// 优化：大幅减少重渲染频率，避免长消息卡死
		shouldRender := false
		
		// 每500个字符渲染一次（从50提高到500），减少90%渲染次数
		respLen := len(m.currentResp)
		if respLen > 0 && respLen%500 == 0 {
			shouldRender = true
		}
		
		// 如果收到思考内容，立即渲染（思考内容通常较短）
		if msg.Reasoning != "" {
			shouldRender = true
		}
		
		// 在句子结束时渲染（提供更好的阅读体验）
		if respLen > 0 {
			lastChar := m.currentResp[respLen-1:]
			if lastChar == "." || lastChar == "!" || lastChar == "?" || lastChar == "\n" {
				shouldRender = true
			}
		}
		
		// 小数据块（可能是最后一块）立即渲染
		if len(msg.Chunk) > 0 && len(msg.Chunk) < 50 {
			shouldRender = true
		}
		
		if shouldRender {
			// 使用优化的渲染方法，只渲染新增内容
			m.renderOptimizedViewport()
		}
		return m, m.checkStream()

	case ToolCallMsg:
		// 收集工具调用，等待流结束后执行
		m.pendingToolCalls = append(m.pendingToolCalls, msg.ToolCalls...)

		// 将工具调用添加到API历史
		m.apiMessages = append(m.apiMessages, api.ToolCallMessage(msg.ToolCalls))

		// 显示工具调用信息
		var toolCallDisplay []string
		for _, toolCall := range msg.ToolCalls {
			toolCallDisplay = append(toolCallDisplay, m.toolManager.FormatToolCallForDisplay(toolCall))
		}

		display := "🔧 AI 请求使用工具:\n" + strings.Join(toolCallDisplay, "\n\n")
		m.messages = append(m.messages, Message{Role: "system", Content: display})

		// 关键修复：工具调用后继续读取流
		return m, tea.Batch(m.updateViewport(), m.checkStream())

	case ToolResultMsg:
		// 显示工具执行结果
		m.messages = append(m.messages, Message{Role: "system", Content: msg.DisplayContent})

		// 将工具结果添加到API历史
		for _, resultMsg := range msg.ResultMessages {
			m.apiMessages = append(m.apiMessages, resultMsg)
		}

		// 清空挂起的工具调用
		m.pendingToolCalls = nil

		// 继续与AI对话（发送工具结果）
		return m, tea.Batch(m.updateViewport(), m.continueStream())

	case StreamErrorMsg:
		m.thinking = false
		errorMsg := fmt.Sprintf("❌ API Error: %v", msg.Error)
		m.messages = append(m.messages, Message{Role: "system", Content: errorMsg})
		return m, m.updateViewport()
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) saveHistory() {
	if len(m.messages) > 0 {
		historyMessages := make([]utils.Message, len(m.messages))
		for i, msg := range m.messages {
			historyMessages[i] = utils.Message{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		utils.SaveHistory(historyMessages)
	}
}

func (m Model) saveChangesToDisk() tea.Cmd {
	return func() tea.Msg {
		if m.editor == nil {
			return ResponseMsg{Content: "编辑系统未初始化"}
		}

		if err := m.editor.SaveToDisk(); err != nil {
			return ResponseMsg{Content: "保存失败: " + err.Error()}
		}

		edits := m.editor.GetCurrentEdits()
		return ResponseMsg{Content: fmt.Sprintf("已保存 %d 个修改到磁盘", len(edits))}
	}
}

func (m Model) View() string {
	if !m.ready {
		return "初始化中..."
	}

	return fmt.Sprintf(
		"%s\n\n%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
		m.helpView(),
	)
}

func (m *Model) updateViewport() tea.Cmd {
	m.viewport.SetContent(m.formatMessages())
	m.viewport.GotoBottom()
	return nil
}

func (m Model) formatMessages() string {
	messageCount := len(m.messages)
	if messageCount == 0 {
		return ""
	}
	
	// 预分配字符串构建器容量，避免多次扩容（初始估算每条消息平均200字符）
	var sb strings.Builder
	sb.Grow(messageCount * 200)
	
	// 限制显示的消息数量，只显示最近的消息
	// 保留最近10条用户消息和对应的AI回复，以及所有系统消息
	const maxUserMessages = 10
	userMessageCount := 0
	
	// 计算需要显示的消息起始位置（从后向前遍历更高效）
	startIndex := 0
	for i := messageCount - 1; i >= 0; i-- {
		if m.messages[i].Role == "user" {
			userMessageCount++
			if userMessageCount > maxUserMessages {
				startIndex = i + 1
				break
			}
		}
	}
	
	// 如果有消息被跳过，显示提示
	if startIndex > 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			fmt.Sprintf("... (显示最近 %d 条对话，共 %d 条) ...\n\n", 
				messageCount-startIndex, messageCount)))
	}
	
	// 渲染从startIndex开始的消息
	for i := startIndex; i < messageCount; i++ {
		msg := m.messages[i]
		switch msg.Role {
		case "user":
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("你: "))
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("AI: "))
			// 直接显示原始内容
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "system":
			// 只显示工具调用、工具结果和错误消息，不显示长的系统提示
			content := msg.Content
			if len(content) < 100 ||
				strings.Contains(content, "🔧") ||
				strings.Contains(content, "✅") ||
				strings.Contains(content, "❌") ||
				strings.Contains(content, "工具执行") ||
							strings.Contains(content, "AI 请求使用工具") {
							sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render("系统: "))
							// 直接显示原始内容
							sb.WriteString(content)
							sb.WriteString("\n\n")			}
		}
	}
	return sb.String()
}

// formatMessagesWithoutLastAssistant 格式化消息但不包含最后一条AI消息（用于流式渲染）
func (m Model) formatMessagesWithoutLastAssistant() string {
	messageCount := len(m.messages)
	if messageCount == 0 {
		return ""
	}
	
	// 如果最后一条是AI消息，则不渲染它
	endIndex := messageCount
	if m.messages[endIndex-1].Role == "assistant" {
		endIndex--
	}
	
	// 如果没有消息需要渲染，返回空
	if endIndex == 0 {
		return ""
	}
	
	// 复用 formatMessages 的逻辑，避免代码重复
	// 创建一个临时消息切片，排除最后一条AI消息
	tempMessages := m.messages[:endIndex]
	
	var sb strings.Builder
	sb.Grow(endIndex * 200)
	
	// 限制显示的消息数量，只显示最近的消息
	const maxUserMessages = 10
	userMessageCount := 0
	
	// 计算需要显示的消息起始位置
	startIndex := 0
	for i := endIndex - 1; i >= 0; i-- {
		if tempMessages[i].Role == "user" {
			userMessageCount++
			if userMessageCount > maxUserMessages {
				startIndex = i + 1
				break
			}
		}
	}
	
	// 如果有消息被跳过，显示提示
	if startIndex > 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
			fmt.Sprintf("... (显示最近 %d 条对话，共 %d 条) ...\n\n", 
				endIndex-startIndex, messageCount)))
	}
	
	// 渲染从startIndex开始的消息
	for i := startIndex; i < endIndex; i++ {
		msg := tempMessages[i]
		switch msg.Role {
		case "user":
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("你: "))
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("AI: "))
			// 直接显示原始内容
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "system":
			content := msg.Content
			if len(content) < 100 ||
				strings.Contains(content, "🔧") ||
				strings.Contains(content, "✅") ||
				strings.Contains(content, "❌") ||
				strings.Contains(content, "工具执行") ||
							strings.Contains(content, "AI 请求使用工具") {
							sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render("系统: "))
							sb.WriteString(content)
							sb.WriteString("\n\n")			}
		}
	}
	return sb.String()
}



// renderOptimizedViewport 优化的视口渲染，只渲染新增内容（增量更新）
func (m *Model) renderOptimizedViewport() {
	// 预分配容量，避免多次扩容（估算：历史消息 + 当前响应 + 思考内容）
	var displayContent strings.Builder
	displayContent.Grow(4096)
	
	// 只在首次或消息完成时渲染历史消息
	if m.renderedLines == nil || len(m.messages) == 0 {
		displayContent.WriteString(m.formatMessagesWithoutLastAssistant())
	} else {
		// 复用已缓存的渲染结果
		for _, line := range m.renderedLines {
			displayContent.WriteString(line)
			displayContent.WriteString("\n")
		}
	}
	
	// 添加思考内容（增量更新）
	if m.currentThink != "" {
		displayContent.WriteString("\n")
		displayContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render("思考: "))
		displayContent.WriteString(m.currentThink)
		displayContent.WriteString("█")
	}
	
	// 添加实时AI响应（增量更新）
	if m.currentResp != "" {
		displayContent.WriteString("\n")
		displayContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("AI: "))
		displayContent.WriteString(m.currentResp)
		displayContent.WriteString("█")
	}
	
	m.viewport.SetContent(displayContent.String())
	m.viewport.GotoBottom()
}

// updateRenderedLinesCache 更新历史消息的渲染缓存
func (m *Model) updateRenderedLinesCache() {
	messageCount := len(m.messages)
	if messageCount == 0 {
		m.renderedLines = nil
		return
	}
	
	// 只缓存最近的消息（避免内存占用过大）
	const maxCacheMessages = 20
	startIndex := 0
	if messageCount > maxCacheMessages {
		startIndex = messageCount - maxCacheMessages
	}
	
	// 预分配容量
	var sb strings.Builder
	sb.Grow(maxCacheMessages * 200)
	
	// 渲染消息到缓存（排除最后一条正在输入的）
	endIndex := messageCount
	if endIndex > 0 && m.messages[endIndex-1].Role == "assistant" && m.thinking {
		endIndex-- // 流式响应时，最后一条AI消息还未完成
	}
	
	for i := startIndex; i < endIndex; i++ {
		msg := m.messages[i]
		switch msg.Role {
		case "user":
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("你: "))
					sb.WriteString(msg.Content)
					sb.WriteString("\n\n")
				case "assistant":
					sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("AI: "))
					// 直接显示原始内容
					sb.WriteString(msg.Content)
					sb.WriteString("\n\n")
				case "system":
					content := msg.Content
					if len(content) < 100 ||
						strings.Contains(content, "🔧") ||
						strings.Contains(content, "✅") ||
						strings.Contains(content, "❌") ||
						strings.Contains(content, "工具执行") ||
						strings.Contains(content, "AI 请求使用工具") {
						sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render("系统: "))
						sb.WriteString(content)
						sb.WriteString("\n\n")
					}
				}	}
	
	// 将渲染结果按行缓存
	content := sb.String()
	if content != "" {
		// 使用高效的字符串分割
		m.renderedLines = strings.Split(strings.TrimRight(content, "\n"), "\n")
	} else {
		m.renderedLines = nil
	}
}

func (m Model) helpView() string {
	help := "Enter: 发送消息 • Ctrl+S: 保存修改 • Esc: 取消思考 • Ctrl+C: 退出"
	if m.thinking {
		help = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("AI正在思考中... ") + "Esc: 取消"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(help)
}

func (m *Model) startStream(input string) tea.Cmd {
	m.thinking = true
	m.currentResp = ""
	m.currentThink = ""

	// 添加用户消息到API历史
	m.apiMessages = append(m.apiMessages, api.TextMessage("user", input))

	// 添加用户消息到界面
	m.messages = append(m.messages, Message{Role: "user", Content: input})

	// 创建统一的API客户端
	client := api.NewClient(m.apiKey)

	// 准备工具
	tools := m.toolManager.GetToolsForAPI()

	// 如果有工具，添加系统提示
	finalMessages := m.apiMessages
	if len(tools) > 0 {
		finalMessages = addSystemPromptIfNeeded(m.apiMessages)
	}

	// 启动流式请求
	m.streamCh, m.reasoningCh, m.toolCallCh, m.streamErrCh = client.StreamChatWithChannel(m.ctx, finalMessages, tools)

	return func() tea.Msg {
		select {
		case chunk := <-m.streamCh:
			if chunk == "" {
				// 流结束
				return CheckStreamMsg{}
			}
			return StreamChunkMsg{Chunk: chunk}
		case reasoning := <-m.reasoningCh:
			return StreamChunkMsg{Reasoning: reasoning}
		case toolCalls := <-m.toolCallCh:
			return ToolCallMsg{ToolCalls: toolCalls}
		case err := <-m.streamErrCh:
			return StreamErrorMsg{Error: err}
		}
	}
}

func (m *Model) checkStream() tea.Cmd {
	return func() tea.Msg {
		select {
		case chunk := <-m.streamCh:
			if chunk == "" {
				// 流结束
				return CheckStreamMsg{}
			}
			return StreamChunkMsg{Chunk: chunk}
		case reasoning := <-m.reasoningCh:
			return StreamChunkMsg{Reasoning: reasoning}
		case toolCalls := <-m.toolCallCh:
			return ToolCallMsg{ToolCalls: toolCalls}
		case err := <-m.streamErrCh:
			return StreamErrorMsg{Error: err}
		}
	}
}

func (m *Model) executePendingTools() tea.Cmd {
	return func() tea.Msg {
		if len(m.pendingToolCalls) == 0 {
			return nil
		}

		// 执行工具调用
		resultMessages, err := m.toolManager.HandleToolCalls(m.pendingToolCalls)
		if err != nil {
			// 创建错误消息
			errorMsg := fmt.Sprintf("工具执行失败: %v", err)
			return ToolResultMsg{
				ResultMessages: []api.Message{api.TextMessage("system", errorMsg)},
				DisplayContent: errorMsg,
			}
		}

		// 格式化显示内容
		var displayContent strings.Builder
		displayContent.WriteString("✅ 工具执行完成:\n")
		for _, msg := range resultMessages {
			if msg.Role == "tool" {
				// 显示工具名称和结果
				toolName := msg.Name
				if toolName == "" {
					toolName = "未知工具"
				}
				displayContent.WriteString(fmt.Sprintf("🔧 %s 结果:\n%s\n\n", toolName, string(msg.Content)))
			}
		}

		return ToolResultMsg{
			ResultMessages: resultMessages,
			DisplayContent: displayContent.String(),
		}
	}
}

func (m *Model) continueStream() tea.Cmd {
	m.thinking = true
	m.currentResp = ""
	m.currentThink = ""

	// 创建统一的API客户端
	client := api.NewClient(m.apiKey)

	// 准备工具
	tools := m.toolManager.GetToolsForAPI()

	// 启动流式请求（使用当前的API历史）
	m.streamCh, m.reasoningCh, m.toolCallCh, m.streamErrCh = client.StreamChatWithChannel(m.ctx, m.apiMessages, tools)

	return func() tea.Msg {
		select {
		case chunk := <-m.streamCh:
			if chunk == "" {
				// 流结束
				return CheckStreamMsg{}
			}
			return StreamChunkMsg{Chunk: chunk}
		case reasoning := <-m.reasoningCh:
			return StreamChunkMsg{Reasoning: reasoning}
		case toolCalls := <-m.toolCallCh:
			return ToolCallMsg{ToolCalls: toolCalls}
		case err := <-m.streamErrCh:
			return StreamErrorMsg{Error: err}
		}
	}
}

// handleCommand 处理命令
func (m *Model) handleCommand(cmd *Command) tea.Cmd {
	switch cmd.Type {
	case CommandTypeClear:
		return m.handleClearCommand()
	case CommandTypeInit:
		return m.handleInitCommand()
	case CommandTypeCheckUpdate:
		return m.handleCheckUpdateCommand()
	case CommandTypeUpdate:
		return m.handleUpdateCommand()
	default:
		// 对于其他命令，显示不支持的消息
		return func() tea.Msg {
			return ResponseMsg{
				Content: fmt.Sprintf("命令 '%s' 暂不支持", FormatCommandType(cmd.Type)),
			}
		}
	}
}

// handleInitCommand 处理 init 命令
func (m *Model) handleInitCommand() tea.Cmd {
	// 发送一个特殊的消息给 AI，让 AI 使用工具来分析项目
	specialMessage := `请分析当前项目并生成 AGENT.md 文件。你可以使用所有可用的工具来：
1. 分析项目结构和文件
2. 读取关键配置文件
3. 理解项目架构和技术栈
4. 生成详细的 AGENT.md 文档

AGENT.md 应该包含：
- 项目概述和用途
- 技术栈和依赖
- 项目结构说明
- 开发约定和最佳实践
- 构建和运行指南
- 注意事项

请使用工具来获取详细信息，然后生成完整的文档。`

	// 将消息添加到对话中
	m.messages = append(m.messages, Message{Role: "user", Content: specialMessage})
	m.textarea.Reset()
	m.thinking = true
	m.currentResp = ""
	m.currentThink = ""

	// 添加到 API 历史
	m.apiMessages = append(m.apiMessages, api.TextMessage("user", specialMessage))

	// 启动流式请求
	client := api.NewClient(m.apiKey)
	tools := m.toolManager.GetToolsForAPI()

	// 如果有工具，添加系统提示
	finalMessages := m.apiMessages
	if len(tools) > 0 {
		finalMessages = addSystemPromptIfNeeded(m.apiMessages)
	}

	m.streamCh, m.reasoningCh, m.toolCallCh, m.streamErrCh = client.StreamChatWithChannel(m.ctx, finalMessages, tools)

	return func() tea.Msg {
		select {
		case chunk := <-m.streamCh:
			if chunk == "" {
				// 流结束
				return CheckStreamMsg{}
			}
			return StreamChunkMsg{Chunk: chunk}
		case reasoning := <-m.reasoningCh:
			return StreamChunkMsg{Reasoning: reasoning}
		case toolCalls := <-m.toolCallCh:
			return ToolCallMsg{ToolCalls: toolCalls}
		case err := <-m.streamErrCh:
			return StreamErrorMsg{Error: err}
		}
	}
}

// handleCheckUpdateCommand 处理检查更新命令
func (m *Model) handleCheckUpdateCommand() tea.Cmd {
	return func() tea.Msg {
		checker := update.NewChecker()
		
		latestVersion, err := checker.GetLatestVersion()
		if err != nil {
			return ResponseMsg{
				Content: fmt.Sprintf("检查更新失败: %v", err),
			}
		}
		
		hasUpdate, _, err := checker.CheckForUpdate(Version)
		if err != nil {
			return ResponseMsg{
				Content: fmt.Sprintf("检查更新失败: %v", err),
			}
		}
		
		if hasUpdate {
			return ResponseMsg{
				Content: fmt.Sprintf("发现新版本!\n当前版本: %s\n最新版本: %s\n\n输入 update 或 /update 开始更新", Version, latestVersion),
			}
		} else {
			return ResponseMsg{
				Content: fmt.Sprintf("当前已是最新版本 (%s)", Version),
			}
		}
	}
}

// handleClearCommand 处理清空命令
func (m *Model) handleClearCommand() tea.Cmd {
	return func() tea.Msg {
		// 清空所有消息
		m.messages = []Message{}
		m.apiMessages = []api.Message{}
		m.currentResp = ""
		m.currentThink = ""
		m.renderedLines = nil
		
		// 取消当前正在进行的操作
		if m.thinking {
			m.thinking = false
			if m.cancel != nil {
				m.cancel()
			}
			// 重新创建context以便下次使用
			m.ctx, m.cancel = context.WithCancel(context.Background())
		}
		
		// 更新视口显示
		m.viewport.SetContent("上下文已清空。可以开始新的对话。\n\n")
		m.viewport.GotoBottom()
		
		return ResponseMsg{
			Content: "上下文和所有消息已清空。",
		}
	}
}

// handleUpdateCommand 处理更新命令
func (m *Model) handleUpdateCommand() tea.Cmd {
	return func() tea.Msg {
		updater := update.NewUpdater()
		
		if err := updater.Update(Version); err != nil {
			return ResponseMsg{
				Content: fmt.Sprintf("更新失败: %v", err),
			}
		}
		
		return ResponseMsg{
			Content: fmt.Sprintf("更新成功! 请重启 PolyAgent 以使用新版本。"),
		}
	}
}

// addSystemPromptIfNeeded 添加系统提示（如果有工具）
func addSystemPromptIfNeeded(messages []api.Message) []api.Message {
	// 检查是否已经有系统提示
	for _, msg := range messages {
		if msg.Role == "system" {
			return messages
		}
	}
	
	// 添加系统提示
	systemPrompt := `你是一个AI助手，可以使用各种工具来帮助用户完成任务。
可用的工具包括：
- 文件操作：读取、写入、搜索文件
- 目录操作：列出目录内容
- Shell命令：执行系统命令
- 网络搜索：搜索网络信息
- Git操作：执行Git命令
- 时间工具：获取当前时间

请根据用户需求选择合适的工具来完成任务。`
	
	result := make([]api.Message, len(messages)+1)
	result[0] = api.TextMessage("system", systemPrompt)
	copy(result[1:], messages)
	
	return result
}
