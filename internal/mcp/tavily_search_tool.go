package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/config"
)

const (
	tavilySearchURL = "https://api.tavily.com/search"
	tavilyTimeout   = 30 * time.Second
)

// TavilySearchTool Tavily 搜索工具
type TavilySearchTool struct {
	Client *http.Client
	APIKey string
}

// NewTavilySearchTool 创建新的 TavilySearchTool 实例
func NewTavilySearchTool() *TavilySearchTool {
	return &TavilySearchTool{
		Client: &http.Client{
			Timeout: tavilyTimeout,
		},
	}
}

func (t *TavilySearchTool) Name() string {
	return "web_search"
}

func (t *TavilySearchTool) Description() string {
	return "使用 Tavily API 进行网页搜索，获取最新、最相关的搜索结果"
}

func (t *TavilySearchTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索关键词或问题",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "返回结果数量 (1-10)",
				"default":     5,
			},
			"search_depth": map[string]interface{}{
				"type":        "string",
				"description": "搜索深度：basic (快速) 或 advanced (深度)",
				"enum":        []string{"basic", "advanced"},
				"default":     "basic",
			},
			"time_range": map[string]interface{}{
				"type":        "string",
				"description": "时间范围：day (一天内), week (一周内), month (一个月内), year (一年内), all (不限时间)",
				"enum":        []string{"day", "week", "month", "year", "all"},
				"default":     "month",
			},
		},
		"required": []string{"query"},
	}
}

// TavilySearchRequest Tavily 搜索请求结构
type TavilySearchRequest struct {
	Query       string `json:"query"`
	MaxResults  int    `json:"max_results,omitempty"`
	SearchDepth string `json:"search_depth,omitempty"`
	TimeRange   string `json:"time_range,omitempty"`
	APIKey      string `json:"api_key"`
}

// TavilySearchResponse Tavily 搜索响应结构
type TavilySearchResponse struct {
	Query   string               `json:"query"`
	Results []TavilySearchResult `json:"results"`
}

// TavilySearchResult 搜索结果项
type TavilySearchResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score,omitempty"`
}

func (t *TavilySearchTool) Execute(args map[string]interface{}) (interface{}, error) {
	// 1. 确保有 API Key
	if err := t.ensureAPIKey(); err != nil {
		return t.getAPIKeyPrompt(), nil
	}

	// 2. 解析参数
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("invalid argument: query is required")
	}

	maxResults := 5
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
		if maxResults < 1 {
			maxResults = 1
		} else if maxResults > 10 {
			maxResults = 10
		}
	}

	searchDepth := "basic"
	if sd, ok := args["search_depth"].(string); ok && (sd == "basic" || sd == "advanced") {
		searchDepth = sd
	}

	timeRange := "month"
	if tr, ok := args["time_range"].(string); ok && (tr == "day" || tr == "week" || tr == "month" || tr == "year" || tr == "all") {
		timeRange = tr
	}

	// 3. 构建请求
	reqBody := TavilySearchRequest{
		Query:       query,
		MaxResults:  maxResults,
		SearchDepth: searchDepth,
		TimeRange:   timeRange,
		APIKey:      t.APIKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), tavilyTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tavilySearchURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 4. 发送请求
	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API error: status %d", resp.StatusCode)
	}

	// 5. 解析响应
	var searchResp TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 6. 格式化结果
	return t.formatResults(query, &searchResp), nil
}

// ensureAPIKey 确保 API Key 已加载
func (t *TavilySearchTool) ensureAPIKey() error {
	if t.APIKey != "" {
		return nil
	}

	// 从配置加载
	key, err := config.GetTavilyAPIKey()
	if err != nil {
		return fmt.Errorf("failed to load API key: %w", err)
	}

	if key == "" {
		return fmt.Errorf("API key not configured")
	}

	t.APIKey = key
	return nil
}

// getAPIKeyPrompt 返回 API Key 配置提示
func (t *TavilySearchTool) getAPIKeyPrompt() string {
	return `# ⚠️ Tavily API Key 未配置

要使用网页搜索功能，需要配置 Tavily API Key。

## 设置步骤：

1. 访问 https://tavily.com/ 注册账号
2. 获取免费 API Key
3. 在配置文件中添加：
   ` + "```yaml" + `
   tavily_api_key: "tvly-xxxxxx"
   ` + "```" + `
   
   配置文件位置：` + "`~/.config/polyagent/config.yaml`" + `

配置完成后，请重新运行搜索。`
}

// formatResults 格式化搜索结果为 Markdown
func (t *TavilySearchTool) formatResults(query string, resp *TavilySearchResponse) string {
	var builder strings.Builder
	builder.Grow(500 + len(resp.Results)*300)

	builder.WriteString(fmt.Sprintf("# 🔍 搜索结果: %q\n\n", query))

	if len(resp.Results) == 0 {
		builder.WriteString("未找到相关结果。\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("找到 %d 个结果：\n\n", len(resp.Results)))

	for i, result := range resp.Results {
		builder.WriteString(fmt.Sprintf("## %d. [%s](%s)\n\n", i+1, escapeMarkdownTitle(result.Title), result.URL))

		if result.Content != "" {
			content := cleanContent(result.Content)
			builder.WriteString(fmt.Sprintf("%s\n\n", content))
		}

		builder.WriteString("---\n\n")
	}

	return builder.String()
}

// escapeMarkdownTitle 转义 Markdown 标题中的特殊字符
func escapeMarkdownTitle(text string) string {
	text = strings.ReplaceAll(text, "[", "\\[")
	text = strings.ReplaceAll(text, "]", "\\]")
	return text
}

// cleanContent 清理并截断内容
func cleanContent(content string) string {
	// 1. 替换换行符为空格
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	// 2. 去除首尾空格
	content = strings.TrimSpace(content)
	// 3. 截断
	runes := []rune(content)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return content
}
