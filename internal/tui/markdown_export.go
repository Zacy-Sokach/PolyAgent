package tui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// 终端能力检测结果
type terminalCapabilities struct {
	supportsColor  bool
	supportsUnicode bool
	width          int
	height         int
}

// 全局终端能力缓存
var (
	terminalCaps *terminalCapabilities
	terminalOnce sync.Once
)

// 渲染缓存项
type renderCacheItem struct {
	content   string
	timestamp time.Time
	hash      string // 内容哈希，用于快速比较
}

// 全局渲染缓存和锁
var (
	renderCache = make(map[string]renderCacheItem)
	cacheMutex  sync.RWMutex
	cacheMaxAge = 10 * time.Minute // 延长缓存时间
	cacheMaxSize = 500             // 增大缓存容量
)

// 简单哈希函数（用于缓存键）
func simpleHash(s string) string {
	if len(s) == 0 {
		return ""
	}
	
	// 使用FNV-1a哈希算法
	hash := uint64(14695981039346656037)
	for i := 0; i < len(s) && i < 1000; i++ { // 只哈希前1000字符，提高性能
		hash ^= uint64(s[i])
		hash *= 1099511628211
	}
	
	// 转换为16进制字符串
	return fmt.Sprintf("%x", hash)
}

// detectTerminalCapabilities 检测终端能力
func detectTerminalCapabilities() *terminalCapabilities {
	caps := &terminalCapabilities{
		supportsColor:   true,  // 默认假设支持颜色
		supportsUnicode: true,  // 默认假设支持Unicode
		width:          80,     // 默认宽度
		height:         24,     // 默认高度
	}

	// 检测环境变量
	if term := os.Getenv("TERM"); term != "" {
		// 检测是否支持颜色
		if strings.Contains(term, "color") || strings.Contains(term, "256") || strings.Contains(term, "xterm") {
			caps.supportsColor = true
		} else if term == "dumb" {
			caps.supportsColor = false
		}

		// 检测终端类型
		if strings.Contains(term, "vt100") || strings.Contains(term, "dumb") {
			caps.supportsUnicode = false
		}
	}

	// 检测COLORTERM环境变量（更可靠的颜色检测）
	if colorterm := os.Getenv("COLORTERM"); colorterm != "" {
		caps.supportsColor = true
	}

	// 检测NO_COLOR环境变量（禁用颜色）
	if os.Getenv("NO_COLOR") != "" {
		caps.supportsColor = false
	}

	// 检测终端大小（如果可能）
	if width := os.Getenv("COLUMNS"); width != "" {
		if w, err := parseTerminalSize(width); err == nil && w > 0 {
			caps.width = w
		}
	}
	if height := os.Getenv("LINES"); height != "" {
		if h, err := parseTerminalSize(height); err == nil && h > 0 {
			caps.height = h
		}
	}

	return caps
}

// parseTerminalSize 解析终端大小字符串
func parseTerminalSize(s string) (int, error) {
	var size int
	_, err := fmt.Sscanf(s, "%d", &size)
	return size, err
}

// getTerminalCapabilities 获取终端能力（单例模式）
func getTerminalCapabilities() *terminalCapabilities {
	terminalOnce.Do(func() {
		terminalCaps = detectTerminalCapabilities()
	})
	return terminalCaps
}

// RenderMarkdownToANSI 将Markdown渲染为ANSI格式的字符串（带缓存和终端兼容性）
func RenderMarkdownToANSI(markdown string) string {
	// 输入验证
	if len(markdown) == 0 {
		return ""
	}

	// 获取终端能力
	caps := getTerminalCapabilities()

	// 如果不支持颜色，创建一个简化版本的渲染器
	if !caps.supportsColor {
		return renderWithoutColors(markdown)
	}

	// 对长内容使用哈希作为缓存键，避免存储完整内容
	contentHash := simpleHash(markdown)
	cacheKey := contentHash
	
	// 对超长内容（>10KB），使用内容长度+前100字符+哈希作为键
	if len(markdown) > 10240 {
		preview := markdown
		if len(preview) > 100 {
			preview = preview[:100]
		}
		cacheKey = fmt.Sprintf("%s_%d_%s", contentHash, len(markdown), simpleHash(preview))
	}

	// 检查缓存
	cacheMutex.RLock()
	if item, exists := renderCache[cacheKey]; exists {
		// 双重检查：确保缓存的内容哈希匹配（防止哈希冲突）
		if time.Since(item.timestamp) < cacheMaxAge && item.hash == contentHash {
			cacheMutex.RUnlock()
			return item.content
		}
	}
	cacheMutex.RUnlock()

	// 缓存未命中或过期，执行渲染
	renderer := NewMarkdownRenderer()
	result := renderer.Render(markdown)

	// 如果不支持Unicode，进行字符替换
	if !caps.supportsUnicode {
		result = replaceUnicodeSymbols(result)
	}

	// 更新缓存（使用更智能的清理策略）
	cacheMutex.Lock()
	
	// 如果缓存过大，清理最旧的20%条目
	if len(renderCache) >= cacheMaxSize {
		// 找出最旧的20%条目
		oldCount := int(float64(cacheMaxSize) * 0.2)
		if oldCount < 1 {
			oldCount = 1
		}
		
		// 创建临时切片排序
		type cacheEntry struct {
			key   string
			item  renderCacheItem
		}
		entries := make([]cacheEntry, 0, len(renderCache))
		for k, v := range renderCache {
			entries = append(entries, cacheEntry{k, v})
		}
		
		// 按时间戳排序（最旧的在前）
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].item.timestamp.After(entries[j].item.timestamp) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
		
		// 删除最旧的条目
		for i := 0; i < oldCount && i < len(entries); i++ {
			delete(renderCache, entries[i].key)
		}
	}
	
	renderCache[cacheKey] = renderCacheItem{
		content:   result,
		timestamp: time.Now(),
		hash:      contentHash,
	}
	cacheMutex.Unlock()

	return result
}

// renderWithoutColors 渲染不带颜色的版本
func renderWithoutColors(markdown string) string {
	// 移除markdown格式，保留纯文本
	lines := strings.Split(markdown, "\n")
	var result strings.Builder
	
	for _, line := range lines {
		// 移除markdown标记
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "`", "")
		line = strings.ReplaceAll(line, "#", "")
		
		result.WriteString(line)
		result.WriteString("\n")
	}
	
	return result.String()
}

// replaceUnicodeSymbols 替换Unicode符号为ASCII替代
func replaceUnicodeSymbols(text string) string {
	replacements := map[string]string{
		"✓": "x",
		"✗": "x",
		"•": "*",
		"→": "->",
		"←": "<-",
		"…": "...",
		"—": "--",
		"\u201C": "\"",
		"\u201D": "\"",
		"\u2018": "'",
		"\u2019": "'",
	}
	
	result := text
	for unicode, ascii := range replacements {
		result = strings.ReplaceAll(result, unicode, ascii)
	}
	
	return result
}

// RenderMarkdownToANSINoCache 不使用缓存的渲染（用于测试或特殊情况）
func RenderMarkdownToANSINoCache(markdown string) string {
	caps := getTerminalCapabilities()
	if !caps.supportsColor {
		return renderWithoutColors(markdown)
	}
	
	renderer := NewMarkdownRenderer()
	result := renderer.Render(markdown)
	
	if !caps.supportsUnicode {
		result = replaceUnicodeSymbols(result)
	}
	
	return result
}

// ClearRenderCache 清空渲染缓存
func ClearRenderCache() {
	cacheMutex.Lock()
	renderCache = make(map[string]renderCacheItem)
	cacheMutex.Unlock()
}
