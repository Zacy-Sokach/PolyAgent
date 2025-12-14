package tui

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderTask 表示一个渲染任务
type RenderTask struct {
	ID       string
	Content  string
	Priority int // 优先级：数字越小优先级越高
	ResultCh chan *RenderResult
}

// RenderResult 表示渲染结果
type RenderResult struct {
	ID      string
	Content string
	Error   error
}

// AsyncRenderer 异步渲染器
type AsyncRenderer struct {
	mdRenderer   *MarkdownRenderer
	taskQueue    chan *RenderTask
	resultCache  *renderResultCache // 缓存渲染结果（有上限，避免无限增长）
	workerPool   int
	shutdown     chan struct{}
	wg           sync.WaitGroup
	batchSize    int
	batchTimeout time.Duration
}

// NewAsyncRenderer 创建新的异步渲染器
func NewAsyncRenderer() *AsyncRenderer {
	ar := &AsyncRenderer{
		mdRenderer:   GetMarkdownRenderer(),
		taskQueue:    make(chan *RenderTask, 20), // 进一步减少队列大小
		resultCache:  newRenderResultCache(128),  // 减少缓存大小，减少内存使用
		shutdown:     make(chan struct{}),
		workerPool:   1,                        // 减少到1个worker，大幅降低CPU使用
		batchSize:    1,                        // 减少批处理大小
		batchTimeout: 2000 * time.Millisecond, // 增加批处理超时到2秒，让更多内容累积
	}

	// 启动worker池
	for i := 0; i < ar.workerPool; i++ {
		ar.wg.Add(1)
		go ar.worker(i)
	}

	return ar
}

// worker 渲染worker
func (ar *AsyncRenderer) worker(_ int) {
	defer ar.wg.Done()

	var batch []*RenderTask
	ticker := time.NewTicker(ar.batchTimeout)
	defer ticker.Stop()

	processBatch := func(tasks []*RenderTask) {
		for _, task := range tasks {
			ar.processTask(task)
		}
	}

	for {
		select {
		case task := <-ar.taskQueue:
			batch = append(batch, task)

			// 达到批处理大小立即处理
			if len(batch) >= ar.batchSize {
				processBatch(batch)
				batch = batch[:0] // 重置切片
			}

		case <-ticker.C:
			// 超时处理当前批次
			if len(batch) > 0 {
				processBatch(batch)
				batch = batch[:0]
			}

		case <-ar.shutdown:
			// 处理剩余任务后退出
			for {
				select {
				case task := <-ar.taskQueue:
					batch = append(batch, task)
				default:
					// 队列已空，处理剩余批次
					if len(batch) > 0 {
						processBatch(batch)
					}
					return
				}
			}
		}
	}
}

// processTask 处理单个渲染任务
func (ar *AsyncRenderer) processTask(task *RenderTask) {
	// 检查缓存
	if result, ok := ar.resultCache.get(task.Content); ok && result != nil {
		select {
		case task.ResultCh <- result:
		case <-time.After(10 * time.Millisecond):
			// 超时，避免阻塞
		}
		return
	}

	// 执行渲染（限制渲染时间）
	done := make(chan struct{})
	var result *RenderResult

	go func() {
		defer func() {
			if r := recover(); r != nil {
				result = &RenderResult{
					ID:      task.ID,
					Content: task.Content, // 返回原始内容
					Error:   fmt.Errorf("渲染错误: %v", r),
				}
			}
			close(done)
		}()

		result = &RenderResult{
			ID:      task.ID,
			Content: ar.mdRenderer.Render(task.Content),
			Error:   nil,
		}
	}()

	// 等待渲染完成或超时
	select {
	case <-done:
		// 缓存结果
		ar.resultCache.add(task.Content, result)

		// 发送结果
		select {
		case task.ResultCh <- result:
		case <-time.After(10 * time.Millisecond):
			// 超时，避免阻塞
		}

	case <-time.After(500 * time.Millisecond):
		// 渲染超时，返回原始内容
		result = &RenderResult{
			ID:      task.ID,
			Content: task.Content,
			Error:   fmt.Errorf("渲染超时"),
		}

		select {
		case task.ResultCh <- result:
		case <-time.After(10 * time.Millisecond):
			// 超时，避免阻塞
		}
	}
}

// RenderAsync 异步渲染内容
func (ar *AsyncRenderer) RenderAsync(id, content string, priority int) <-chan *RenderResult {
	resultCh := make(chan *RenderResult, 1)

	task := &RenderTask{
		ID:       id,
		Content:  content,
		Priority: priority,
		ResultCh: resultCh,
	}

	select {
	case ar.taskQueue <- task:
	default:
		// 队列已满，返回空结果
		resultCh <- &RenderResult{
			ID:      id,
			Content: content, // 返回原始内容
			Error:   nil,
		}
	}

	return resultCh
}

// RenderSync 同步渲染内容（带缓存）
func (ar *AsyncRenderer) RenderSync(content string) string {
	// 检查缓存
	if result, ok := ar.resultCache.get(content); ok && result != nil {
		return result.Content
	}

	// 直接渲染并缓存
	rendered := ar.mdRenderer.Render(content)
	result := &RenderResult{
		Content: rendered,
		Error:   nil,
	}
	ar.resultCache.add(content, result)

	return rendered
}

// ClearCache 清空渲染缓存
func (ar *AsyncRenderer) ClearCache() {
	ar.resultCache.clear()
}

// Shutdown 关闭异步渲染器
func (ar *AsyncRenderer) Shutdown() {
	close(ar.shutdown)
	ar.wg.Wait()
}

// renderResultCache 一个简单的LRU缓存，限制渲染结果数量，避免无限增长
type renderResultCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type cacheItem struct {
	key   string
	value *RenderResult
}

func newRenderResultCache(capacity int) *renderResultCache {
	if capacity <= 0 {
		capacity = 256
	}
	return &renderResultCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *renderResultCache) get(key string) (*RenderResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		item := elem.Value.(cacheItem)
		return item.value, true
	}
	return nil, false
}

func (c *renderResultCache) add(key string, value *RenderResult) {
	if value == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		elem.Value = cacheItem{key: key, value: value}
		c.order.MoveToFront(elem)
		return
	}

	elem := c.order.PushFront(cacheItem{key: key, value: value})
	c.items[key] = elem

	if c.order.Len() > c.capacity {
		if tail := c.order.Back(); tail != nil {
			item := tail.Value.(cacheItem)
			delete(c.items, item.key)
			c.order.Remove(tail)
		}
	}
}

func (c *renderResultCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element, c.capacity)
	c.order.Init()
}

// RenderPipeline 渲染管道：按行增量解析并输出 ANSI 字符串
type RenderPipeline struct {
	inputCh  chan string
	outputCh chan string
	shutdown chan struct{}

	// 状态机
	pending     strings.Builder // 尾部未成行
	inCodeBlock bool

	// 样式（预构建，避免重复创建）
	headingStyle lipgloss.Style
	codeStyle    lipgloss.Style
	listStyle    lipgloss.Style
	textStyle    lipgloss.Style
}

// NewRenderPipeline 创建渲染管道
func NewRenderPipeline(_ *AsyncRenderer) *RenderPipeline {
	rp := &RenderPipeline{
		inputCh:   make(chan string, 32),
		outputCh:  make(chan string, 64),
		shutdown:  make(chan struct{}),
		codeStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("236")),
		headingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),
		listStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("78")),
		textStyle: lipgloss.NewStyle(),
	}

	go rp.process()
	return rp
}

// process 处理渲染管道
func (rp *RenderPipeline) process() {
	for {
		select {
		case chunk, ok := <-rp.inputCh:
			if !ok {
				rp.flushPending()
				close(rp.outputCh)
				return
			}
			rp.handleChunk(chunk)
		case <-rp.shutdown:
			rp.flushPending()
			close(rp.outputCh)
			return
		}
	}
}

// handleChunk 追加 chunk 并输出完整行
func (rp *RenderPipeline) handleChunk(chunk string) {
	rp.pending.WriteString(chunk)
	text := rp.pending.String()
	rp.pending.Reset()

	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			line := text[start : i+1]
			rp.outputLine(line)
			start = i + 1
		}
	}

	// 剩余部分保留
	if start < len(text) {
		rp.pending.WriteString(text[start:])
	}
}

// outputLine 将单行通过状态机转换为 ANSI
func (rp *RenderPipeline) outputLine(line string) {
	rendered := rp.parseLine(line)
	select {
	case rp.outputCh <- rendered:
		// sent successfully
	case <-rp.shutdown:
		// shutting down, drop
	default:
		// 输出通道已满时丢弃，避免阻塞导致界面卡死
	}
}

// parseLine 基于简单状态机的增量解析
func (rp *RenderPipeline) parseLine(line string) string {
	trimmed := strings.TrimSpace(line)

	// 代码块切换
	if strings.HasPrefix(trimmed, "```") {
		rp.inCodeBlock = !rp.inCodeBlock
		return rp.codeStyle.Render(trimmed) + "\n"
	}

	if rp.inCodeBlock {
		return rp.codeStyle.Render(strings.TrimRight(line, "\n")) + "\n"
	}

	// 标题
	if strings.HasPrefix(trimmed, "#") {
		return rp.headingStyle.Render(trimmed) + "\n"
	}

	// 列表
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return rp.listStyle.Render(trimmed) + "\n"
	}

	// 行内粗体（简单替换）
	if strings.Contains(line, "**") {
		rendered := replaceBold(line, rp.textStyle)
		return rendered
	}

	return rp.textStyle.Render(line)
}

// flushPending 输出剩余尾部
func (rp *RenderPipeline) flushPending() {
	if rp.pending.Len() == 0 {
		return
	}
	line := rp.pending.String()
	rp.pending.Reset()
	rp.outputLine(line)
}

// AddChunk 添加内容块
func (rp *RenderPipeline) AddChunk(chunk string) {
	select {
	case rp.inputCh <- chunk:
	case <-rp.shutdown:
	default:
		// 输入通道满时丢弃当前增量，避免阻塞主循环
	}
}

// GetOutput 获取渲染输出
func (rp *RenderPipeline) GetOutput() <-chan string {
	return rp.outputCh
}

// Close 关闭渲染管道
func (rp *RenderPipeline) Close() {
	select {
	case <-rp.shutdown:
		// 已关闭
	default:
		close(rp.shutdown)
		close(rp.inputCh)
	}
}

// replaceBold 做最简单的粗体替换，避免使用复杂正则
func replaceBold(s string, base lipgloss.Style) string {
	result := strings.Builder{}
	result.Grow(len(s))

	inBold := false
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "**") {
			inBold = !inBold
			i += 2
			continue
		}
		ch := s[i]
		if inBold {
			result.WriteString(base.Bold(true).Render(string(ch)))
		} else {
			result.WriteByte(ch)
		}
		i++
	}
	return result.String()
}
