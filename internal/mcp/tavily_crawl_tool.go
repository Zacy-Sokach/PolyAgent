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
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

const (
	tavilyCrawlURL = "https://api.tavily.com/crawl"
	crawlTimeout   = 150 * time.Second // Crawl 可能需要更长时间
)

// TavilyCrawlTool Tavily 爬取工具
type TavilyCrawlTool struct {
	Client *http.Client
	APIKey string
}

// NewTavilyCrawlTool 创建新的 TavilyCrawlTool 实例
func NewTavilyCrawlTool() *TavilyCrawlTool {
	return &TavilyCrawlTool{
		Client: &http.Client{
			Timeout: crawlTimeout,
		},
	}
}

func (t *TavilyCrawlTool) Name() string {
	return "web_crawl"
}

func (t *TavilyCrawlTool) Description() string {
	return "深度爬取网站内容，提取多个页面的结构化信息。适合获取完整文档或多页面内容"
}

func (t *TavilyCrawlTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"base_url": map[string]interface{}{
				"type":        "string",
				"description": "起始 URL，爬取的根地址",
			},
			"max_depth": map[string]interface{}{
				"type":        "integer",
				"description": "爬取深度，从起始页面开始的最大层级 (默认 2)",
				"default":     2,
			},
			"max_links_per_level": map[string]interface{}{
				"type":        "integer",
				"description": "每层最大链接数 (默认 10)",
				"default":     10,
			},
			"total_max_links": map[string]interface{}{
				"type":        "integer",
				"description": "总最大链接数 (默认 50)",
				"default":     50,
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "输出格式：markdown 或 text",
				"enum":        []string{"markdown", "text"},
				"default":     "markdown",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "超时时间（秒），范围 10-150",
				"default":     60,
			},
			"include_patterns": map[string]interface{}{
				"type":        "array",
				"description": "URL 包含正则表达式列表（可选）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"exclude_patterns": map[string]interface{}{
				"type":        "array",
				"description": "URL 排除正则表达式列表（可选）",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"base_url"},
	}
}

// TavilyCrawlRequest Tavily 爬取请求结构
type TavilyCrawlRequest struct {
	BaseURL          string   `json:"base_url"`
	MaxDepth         int      `json:"max_depth,omitempty"`
	MaxLinksPerLevel int      `json:"max_links_per_level,omitempty"`
	TotalMaxLinks    int      `json:"total_max_links,omitempty"`
	Format           string   `json:"format,omitempty"`
	Timeout          int      `json:"timeout,omitempty"`
	IncludePatterns  []string `json:"include_patterns,omitempty"`
	ExcludePatterns  []string `json:"exclude_patterns,omitempty"`
	APIKey           string   `json:"api_key"`
}

// TavilyCrawlResponse Tavily 爬取响应结构
type TavilyCrawlResponse struct {
	BaseURL string              `json:"base_url"`
	Results []TavilyCrawlResult `json:"results"`
}

// TavilyCrawlResult 爬取结果项
type TavilyCrawlResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

func (t *TavilyCrawlTool) Execute(args map[string]interface{}) (interface{}, error) {
	// 1. 确保有 API Key
	if err := t.ensureAPIKey(); err != nil {
		return t.getAPIKeyPrompt(), nil
	}

	// 2. 解析参数
	baseURL, ok := args["base_url"].(string)
	if !ok || strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("invalid argument: base_url is required")
	}

	maxDepth := getIntArg(args, "max_depth", 2)
	maxLinksPerLevel := getIntArg(args, "max_links_per_level", 10)
	totalMaxLinks := getIntArg(args, "total_max_links", 50)
	timeout := getIntArg(args, "timeout", 60)

	if timeout < 10 {
		timeout = 10
	} else if timeout > 150 {
		timeout = 150
	}

	format := "markdown"
	if f, ok := args["format"].(string); ok && (f == "markdown" || f == "text") {
		format = f
	}

	var includePatterns, excludePatterns []string
	if patterns, ok := args["include_patterns"].([]interface{}); ok {
		includePatterns = toStringSlice(patterns)
	}
	if patterns, ok := args["exclude_patterns"].([]interface{}); ok {
		excludePatterns = toStringSlice(patterns)
	}

	// 3. 构建请求
	reqBody := TavilyCrawlRequest{
		BaseURL:          baseURL,
		MaxDepth:         maxDepth,
		MaxLinksPerLevel: maxLinksPerLevel,
		TotalMaxLinks:    totalMaxLinks,
		Format:           format,
		Timeout:          timeout,
		IncludePatterns:  includePatterns,
		ExcludePatterns:  excludePatterns,
		APIKey:           t.APIKey,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+10)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tavilyCrawlURL, bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("crawl API error: status %d", resp.StatusCode)
	}

	// 5. 解析响应
	var crawlResp TavilyCrawlResponse
	if err := json.NewDecoder(resp.Body).Decode(&crawlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 6. 格式化结果
	return t.formatResults(baseURL, &crawlResp), nil
}

// ensureAPIKey 确保 API Key 已加载
func (t *TavilyCrawlTool) ensureAPIKey() error {
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
func (t *TavilyCrawlTool) getAPIKeyPrompt() string {
	return `# ⚠️ Tavily API Key 未配置

要使用网站爬取功能，需要配置 Tavily API Key。

## 设置步骤：

1. 访问 https://tavily.com/ 注册账号
2. 获取免费 API Key
3. 在配置文件中添加：
   ` + "```yaml" + `
   tavily_api_key: "tvly-xxxxxx"
   ` + "```" + `
   
   配置文件位置：` + "`" + utils.GetConfigPathForDisplay() + "`" + `

配置完成后，请重新运行爬取。`
}

// formatResults 格式化爬取结果
func (t *TavilyCrawlTool) formatResults(baseURL string, resp *TavilyCrawlResponse) string {
	var builder strings.Builder
	builder.Grow(1000 + len(resp.Results)*500)

	builder.WriteString(fmt.Sprintf("# 🕷️ 网站爬取结果: %s\n\n", baseURL))

	if len(resp.Results) == 0 {
		builder.WriteString("未爬取到任何内容。\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("爬取了 %d 个页面：\n\n", len(resp.Results)))

	for i, result := range resp.Results {
		builder.WriteString(fmt.Sprintf("## 页面 %d: %s\n\n", i+1, result.URL))

		if result.Content != "" {
			// 内容已经是 markdown 或 text 格式
			builder.WriteString(result.Content)
			builder.WriteString("\n\n")
		}

		builder.WriteString("---\n\n")
	}

	return builder.String()
}

// toStringSlice 将 []interface{} 转换为 []string
func toStringSlice(arr []interface{}) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// getIntArg 安全获取 int 参数
func getIntArg(args map[string]interface{}, key string, fallback int) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
	}
	return fallback
}
