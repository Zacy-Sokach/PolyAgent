package tui

import (
	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

// MessageManager 管理消息和对话状态
type MessageManager struct {
	messages      []Message
	apiMessages   []api.Message
	maxMessages   int
}

// NewMessageManager 创建新的消息管理器
func NewMessageManager(maxMessages int) *MessageManager {
	return &MessageManager{
		messages:    []Message{},
		apiMessages: []api.Message{},
		maxMessages: maxMessages,
	}
}

// AddMessage 添加消息到消息列表
func (m *MessageManager) AddMessage(role, content string) {
	m.messages = append(m.messages, Message{Role: role, Content: content})
	m.limitMessages()
}

// AddAPIMessage 添加API消息到API消息列表
func (m *MessageManager) AddAPIMessage(msg api.Message) {
	m.apiMessages = append(m.apiMessages, msg)
}

// GetMessages 获取消息列表
func (m *MessageManager) GetMessages() []Message {
	return m.messages
}

// GetAPIMessages 获取API消息列表
func (m *MessageManager) GetAPIMessages() []api.Message {
	return m.apiMessages
}

// ClearMessages 清空消息列表
func (m *MessageManager) ClearMessages() {
	m.messages = []Message{}
	m.apiMessages = []api.Message{}
}

// limitMessages 限制消息数量
func (m *MessageManager) limitMessages() {
	if len(m.messages) > m.maxMessages {
		// 保留最新的消息，移除最旧的
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}
}

// SaveHistory 保存历史记录
func (m *MessageManager) SaveHistory() {
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