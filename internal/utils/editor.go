package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EditOperation 原子编辑操作
type EditOperation struct {
	Type      string // "insert", "delete"
	FilePath  string // 文件路径
	Offset    int    // 字符偏移量
	Length    int    // 删除长度（delete时）
	Content   string // 插入内容（insert时）
	Timestamp time.Time
}

// SessionMarker 轻量级会话标记
type SessionMarker struct {
	ID         string
	Timestamp  time.Time
	FileHashes map[string]string // 文件路径 -> 初始哈希
}

// TextBuffer 内存文本缓冲区
type TextBuffer struct {
	Content string
}

// FileState 文件状态
type FileState struct {
	Path   string
	Buffer *TextBuffer
	Hash   string
}

// Editor 编辑系统
type Editor struct {
	currentSession *SessionMarker
	sessionEdits   []EditOperation
	fileStates     map[string]*FileState
}

// NewEditor 创建新的编辑系统
func NewEditor() *Editor {
	return &Editor{
		fileStates: make(map[string]*FileState),
	}
}

// StartSession 开始新会话
func (e *Editor) StartSession() error {
	if e.currentSession != nil {
		return fmt.Errorf("已有活跃会话，请先结束当前会话")
	}

	// 获取当前目录所有代码文件
	files, err := e.getCodeFiles()
	if err != nil {
		return fmt.Errorf("获取代码文件失败: %w", err)
	}

	// 创建会话标记
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	fileHashes := make(map[string]string)

	// 初始化文件状态并计算哈希
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue // 跳过无法读取的文件
		}

		hash := e.calculateHash(string(content))
		fileHashes[file] = hash

		e.fileStates[file] = &FileState{
			Path:   file,
			Buffer: &TextBuffer{Content: string(content)},
			Hash:   hash,
		}
	}

	e.currentSession = &SessionMarker{
		ID:         sessionID,
		Timestamp:  time.Now(),
		FileHashes: fileHashes,
	}
	e.sessionEdits = nil

	return nil
}

// EndSession 结束当前会话
func (e *Editor) EndSession() {
	e.currentSession = nil
	e.sessionEdits = nil
	// 保留 fileStates 供下次会话使用
}

// InsertText 插入文本
func (e *Editor) InsertText(filePath string, offset int, content string) error {
	state, ok := e.fileStates[filePath]
	if !ok {
		// 如果文件不在状态中，先加载
		if err := e.loadFile(filePath); err != nil {
			return err
		}
		state = e.fileStates[filePath]
	}

	// 验证偏移量
	if offset < 0 || offset > len(state.Buffer.Content) {
		return fmt.Errorf("偏移量 %d 超出范围 (0-%d)", offset, len(state.Buffer.Content))
	}

	// 执行插入
	oldContent := state.Buffer.Content
	state.Buffer.Content = oldContent[:offset] + content + oldContent[offset:]

	// 记录操作
	e.sessionEdits = append(e.sessionEdits, EditOperation{
		Type:      "insert",
		FilePath:  filePath,
		Offset:    offset,
		Length:    0,
		Content:   content,
		Timestamp: time.Now(),
	})

	return nil
}

// DeleteText 删除文本
func (e *Editor) DeleteText(filePath string, offset int, length int) error {
	state, ok := e.fileStates[filePath]
	if !ok {
		return fmt.Errorf("文件未加载: %s", filePath)
	}

	// 验证偏移量和长度
	if offset < 0 || offset+length > len(state.Buffer.Content) {
		return fmt.Errorf("删除范围超出文件边界")
	}

	// 获取被删除的内容
	deletedContent := state.Buffer.Content[offset : offset+length]

	// 执行删除
	state.Buffer.Content = state.Buffer.Content[:offset] + state.Buffer.Content[offset+length:]

	// 记录操作
	e.sessionEdits = append(e.sessionEdits, EditOperation{
		Type:      "delete",
		FilePath:  filePath,
		Offset:    offset,
		Length:    length,
		Content:   deletedContent,
		Timestamp: time.Now(),
	})

	return nil
}

// ReplaceText 替换文本（插入+删除的组合）
func (e *Editor) ReplaceText(filePath string, offset int, length int, newContent string) error {
	// 先删除旧内容
	if err := e.DeleteText(filePath, offset, length); err != nil {
		return err
	}
	// 再插入新内容
	if err := e.InsertText(filePath, offset, newContent); err != nil {
		return err
	}
	return nil
}

// RollbackSession 回退当前会话的所有修改
func (e *Editor) RollbackSession() error {
	if e.currentSession == nil {
		return fmt.Errorf("没有活跃会话")
	}

	// 逆序执行反向操作
	for i := len(e.sessionEdits) - 1; i >= 0; i-- {
		op := e.sessionEdits[i]
		if err := e.applyInverseOperation(op); err != nil {
			return fmt.Errorf("回退操作失败 (操作 %d): %w", i, err)
		}
	}

	// 验证文件哈希
	for filePath, expectedHash := range e.currentSession.FileHashes {
		state, ok := e.fileStates[filePath]
		if !ok {
			continue // 文件可能已被删除
		}

		currentHash := e.calculateHash(state.Buffer.Content)
		if currentHash != expectedHash {
			return fmt.Errorf("文件 %s 哈希不匹配，可能已被外部修改", filePath)
		}
	}

	// 清空编辑记录
	e.sessionEdits = nil

	return nil
}

// SaveToDisk 将内存中的修改保存到磁盘
func (e *Editor) SaveToDisk() error {
	for _, state := range e.fileStates {
		if err := os.WriteFile(state.Path, []byte(state.Buffer.Content), 0644); err != nil {
			return fmt.Errorf("保存文件 %s 失败: %w", state.Path, err)
		}
	}
	return nil
}

// GetCurrentEdits 获取当前会话的编辑记录
func (e *Editor) GetCurrentEdits() []EditOperation {
	return e.sessionEdits
}

// GetFileContent 获取文件当前内容
func (e *Editor) GetFileContent(filePath string) (string, error) {
	state, ok := e.fileStates[filePath]
	if !ok {
		return "", fmt.Errorf("文件未加载: %s", filePath)
	}
	return state.Buffer.Content, nil
}

// LoadFile 加载文件到编辑器
func (e *Editor) LoadFile(filePath string) error {
	return e.loadFile(filePath)
}

// 辅助方法

func (e *Editor) getCodeFiles() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var files []string
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && isCodeFile(filepath.Ext(path)) {
			relPath, err := filepath.Rel(cwd, path)
			if err == nil {
				files = append(files, relPath)
			}
		}
		return nil
	})

	return files, err
}

func (e *Editor) loadFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	hash := e.calculateHash(string(content))
	e.fileStates[filePath] = &FileState{
		Path:   filePath,
		Buffer: &TextBuffer{Content: string(content)},
		Hash:   hash,
	}

	return nil
}

func (e *Editor) calculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (e *Editor) applyInverseOperation(op EditOperation) error {
	switch op.Type {
	case "insert":
		// 插入的反向操作是删除
		return e.DeleteText(op.FilePath, op.Offset, len(op.Content))
	case "delete":
		// 删除的反向操作是插入
		return e.InsertText(op.FilePath, op.Offset, op.Content)
	default:
		return fmt.Errorf("未知操作类型: %s", op.Type)
	}
}
