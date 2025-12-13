package mcp

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileEngine 统一的文件操作引擎
type FileEngine struct {
	cache  *fileCache
	config *FileEngineConfig
}

// FileEngineConfig 文件引擎配置
type FileEngineConfig struct {
	// 路径白名单（限制在项目目录内）
	AllowedRoots []string
	// 文件类型黑名单
	BlacklistedExts []string
	// 最大文件大小（字节）
	MaxFileSize int64
	// 是否启用缓存
	EnableCache bool
	// 备份目录
	BackupDir string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *FileEngineConfig {
	return &FileEngineConfig{
		AllowedRoots:    []string{"."},
		BlacklistedExts: []string{".exe", ".dll", ".so", ".dylib", ".bin"},
		MaxFileSize:     10 * 1024 * 1024, // 10MB
		EnableCache:     true,
		BackupDir:       ".polyagent-backups",
	}
}

// NewFileEngine 创建文件操作引擎
func NewFileEngine(config *FileEngineConfig) *FileEngine {
	if config == nil {
		config = DefaultConfig()
	}
	
	engine := &FileEngine{
		config: config,
	}
	
	if config.EnableCache {
		engine.cache = newFileCache()
	}
	
	return engine
}

// ValidatePath 验证路径是否允许访问
func (e *FileEngine) ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	
	// 解析符号链接，防止路径遍历
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to evaluate symlinks: %w", err)
	}
	if err == nil {
		absPath = realPath
	}
	
	// 检查是否在允许的根目录内
	allowed := false
	for _, root := range e.config.AllowedRoots {
		absRoot, _ := filepath.Abs(root)
		if strings.HasPrefix(absPath, absRoot) {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return fmt.Errorf("path outside allowed roots: %s", path)
	}
	
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(absPath))
	for _, blacklisted := range e.config.BlacklistedExts {
		if ext == blacklisted {
			return fmt.Errorf("file type not allowed: %s", ext)
		}
	}
	
	return nil
}

// ReadFile 读取文件内容（带缓存）
func (e *FileEngine) ReadFile(path string, forceRefresh bool) ([]byte, error) {
	if err := e.ValidatePath(path); err != nil {
		return nil, err
	}
	
	// 检查缓存（如果未强制刷新）
	if !forceRefresh && e.cache != nil {
		if content, hit := e.cache.get(path); hit {
			return content, nil
		}
	}
	
	// 检查文件大小
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	if info.Size() > e.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %s (%.2f MB)", path, float64(info.Size())/1024/1024)
	}
	
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	// 写入缓存
	if e.cache != nil {
		e.cache.set(path, content)
	}
	
	return content, nil
}

// WriteFile 写入文件（带备份）
func (e *FileEngine) WriteFile(path string, content []byte, backup bool) error {
	if err := e.ValidatePath(path); err != nil {
		return err
	}
	
	// 创建备份
	if backup {
		if err := e.createBackup(path); err != nil {
			return fmt.Errorf("创建备份失败: %w", err)
		}
	}
	
	// 使用临时文件保证原子性
	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		return err
	}
	
	// 原子替换
	if err := os.Rename(tempFile, path); err != nil {
		os.Remove(tempFile) // 清理临时文件
		return err
	}
	
	// 更新缓存
	if e.cache != nil {
		e.cache.set(path, content)
	}
	
	return nil
}

// createBackup 创建文件备份
func (e *FileEngine) createBackup(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，无需备份
		}
		return err
	}
	
	// 创建备份目录
	backupDir := e.config.BackupDir
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}
	
	// 生成备份文件名
	hash := sha256.Sum256([]byte(path))
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%x-%s.backup", 
		filepath.Base(path), hash[:8], timestamp)
	backupPath := filepath.Join(backupDir, backupName)
	
	return os.WriteFile(backupPath, content, 0644)
}

// FileWalker 文件遍历器
type FileWalker struct {
	engine      *FileEngine
	root        string
	include     string
	exclude     string
	maxDepth    int
	currentDepth int
}

// NewFileWalker 创建文件遍历器
func (e *FileEngine) NewFileWalker(root, include, exclude string) *FileWalker {
	return &FileWalker{
		engine:   e,
		root:     root,
		include:  include,
		exclude:  exclude,
		maxDepth: -1, // 无限制
	}
}

// SetMaxDepth 设置最大遍历深度
func (w *FileWalker) SetMaxDepth(depth int) {
	w.maxDepth = depth
}

// Walk 遍历文件并执行回调
func (w *FileWalker) Walk(fn func(path string, info fs.FileInfo) error) error {
	return filepath.Walk(w.root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 深度检查
		if w.maxDepth >= 0 {
			relPath, _ := filepath.Rel(w.root, path)
			depth := strings.Count(relPath, string(os.PathSeparator))
			if info.IsDir() {
				depth-- // 目录本身不计入深度
			}
			if depth > w.maxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		
		// 跳过目录
		if info.IsDir() {
			return nil
		}
		
		// 验证路径
		if err := w.engine.ValidatePath(path); err != nil {
			return nil // 跳过不允许访问的文件
		}
		
		// 应用包含模式
		if w.include != "" && w.include != "*" {
			matched, err := filepath.Match(w.include, filepath.Base(path))
			if err != nil || !matched {
				return nil
			}
		}
		
		// 应用排除模式
		if w.exclude != "" {
			matched, err := filepath.Match(w.exclude, filepath.Base(path))
			if err == nil && matched {
				return nil
			}
		}
		
		return fn(path, info)
	})
}

// fileCache 文件内容缓存
type fileCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	maxSize int
}

type cacheItem struct {
	content []byte
	time    time.Time
}

func newFileCache() *fileCache {
	return &fileCache{
		items:   make(map[string]*cacheItem),
		maxSize: 100, // 最多缓存100个文件
	}
}

func (c *fileCache) get(path string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, ok := c.items[path]
	if !ok {
		return nil, false
	}
	
	// 检查是否过期（5分钟）
	if time.Since(item.time) > 5*time.Minute {
		return nil, false
	}
	
	return item.content, true
}

func (c *fileCache) set(path string, content []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 清理旧缓存
	if len(c.items) >= c.maxSize {
		c.cleanup()
	}
	
	c.items[path] = &cacheItem{
		content: content,
		time:    time.Now(),
	}
}

func (c *fileCache) cleanup() {
	// LRU 策略：删除最旧的缓存项，直到数量降到 maxSize 的 50%
	type itemWithPath struct {
		path string
		item *cacheItem
	}
	
	itemCount := len(c.items)
	targetSize := c.maxSize / 2
	
	// 如果不需要清理，直接返回
	if itemCount <= targetSize {
		return
	}
	
	// 只创建需要大小的切片（避免过度分配）
	items := make([]itemWithPath, 0, itemCount)
	for path, item := range c.items {
		items = append(items, itemWithPath{path, item})
	}
	
	// 使用高效的排序算法（按时间升序排序，旧的在前）
	// Go 的 sort.Slice 使用快速排序，平均 O(n log n)
	sort.Slice(items, func(i, j int) bool {
		return items[i].item.time.Before(items[j].item.time)
	})
	
	// 删除前 50%（最旧的）
	deleteCount := itemCount - targetSize
	for i := 0; i < deleteCount; i++ {
		delete(c.items, items[i].path)
	}
}

// ClearCache 清空缓存
func (e *FileEngine) ClearCache() {
	if e.cache != nil {
		e.cache.mu.Lock()
		defer e.cache.mu.Unlock()
		e.cache.items = make(map[string]*cacheItem)
	}
}
