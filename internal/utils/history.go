package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func SaveHistory(messages []Message) error {
	historyPath, err := getHistoryPath()
	if err != nil {
		return fmt.Errorf("获取历史文件路径失败: %w", err)
	}

	entry := HistoryEntry{
		Timestamp: time.Now(),
		Messages:  messages,
	}

	var history []HistoryEntry

	if _, err := os.Stat(historyPath); err == nil {
		data, err := os.ReadFile(historyPath)
		if err == nil {
			json.Unmarshal(data, &history)
		}
	}

	history = append(history, entry)

	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化历史失败: %w", err)
	}

	historyDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("创建历史目录失败: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return fmt.Errorf("写入历史文件失败: %w", err)
	}

	return nil
}

func LoadHistory() ([]HistoryEntry, error) {
	historyPath, err := getHistoryPath()
	if err != nil {
		return nil, fmt.Errorf("获取历史文件路径失败: %w", err)
	}

	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return []HistoryEntry{}, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("读取历史文件失败: %w", err)
	}

	var history []HistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("解析历史文件失败: %w", err)
	}

	return history, nil
}

func getHistoryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(homeDir, ".config", "polyagent", "history.json"), nil
}
