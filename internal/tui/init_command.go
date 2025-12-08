package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// analyzeProjectAndCreateAgentMD 分析项目并创建/更新 AGENT.md 文件
func analyzeProjectAndCreateAgentMD() (string, error) {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %v", err)
	}

	// 分析项目基本信息
	projectInfo, err := analyzeProject(cwd)
	if err != nil {
		return "", fmt.Errorf("分析项目失败: %v", err)
	}

	// 生成 AGENT.md 内容
	agentMDContent := generateAgentMD(projectInfo)

	// 写入 AGENT.md 文件
	agentMDPath := filepath.Join(cwd, "AGENT.md")
	if err := os.WriteFile(agentMDPath, []byte(agentMDContent), 0644); err != nil {
		return "", fmt.Errorf("写入 AGENT.md 失败: %v", err)
	}

	return fmt.Sprintf("✅ 项目分析完成，AGENT.md 已创建/更新\n\n%s", projectInfo.Summary), nil
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Name         string
	Path         string
	Type         string
	MainLanguage string
	Languages    []string
	Dependencies map[string][]string
	Structure    []string
	Summary      string
}

// analyzeProject 分析项目结构
func analyzeProject(projectPath string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Name:         filepath.Base(projectPath),
		Path:         projectPath,
		Dependencies: make(map[string][]string),
		Languages:    []string{},
		Structure:    []string{},
	}

	// 分析项目类型和语言
	if err := analyzeProjectTypeAndLanguage(info); err != nil {
		return nil, err
	}

	// 分析依赖
	if err := analyzeDependencies(info); err != nil {
		return nil, err
	}

	// 分析目录结构
	if err := analyzeDirectoryStructure(info); err != nil {
		return nil, err
	}

	// 生成摘要
	info.Summary = generateProjectSummary(info)

	return info, nil
}

// analyzeProjectTypeAndLanguage 分析项目类型和编程语言
func analyzeProjectTypeAndLanguage(info *ProjectInfo) error {
	// 检查常见的项目配置文件
	files, err := os.ReadDir(info.Path)
	if err != nil {
		return err
	}

	// 检查文件扩展名以确定主要语言
	extCount := make(map[string]int)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != "" {
			extCount[ext]++
		}

		// 检查项目配置文件
		switch file.Name() {
		case "go.mod":
			info.Type = "Go 项目"
			info.MainLanguage = "Go"
		case "package.json":
			info.Type = "Node.js 项目"
			info.MainLanguage = "JavaScript/TypeScript"
		case "Cargo.toml":
			info.Type = "Rust 项目"
			info.MainLanguage = "Rust"
		case "pyproject.toml", "requirements.txt":
			info.Type = "Python 项目"
			info.MainLanguage = "Python"
		case "pom.xml", "build.gradle", "build.gradle.kts":
			info.Type = "Java 项目"
			info.MainLanguage = "Java"
		case "composer.json":
			info.Type = "PHP 项目"
			info.MainLanguage = "PHP"
		case "Gemfile":
			info.Type = "Ruby 项目"
			info.MainLanguage = "Ruby"
		}
	}

	// 如果没有检测到项目类型，根据文件扩展名推断
	if info.Type == "" {
		// 找出最多的文件扩展名
		maxCount := 0
		mainExt := ""
		for ext, count := range extCount {
			if count > maxCount {
				maxCount = count
				mainExt = ext
			}
		}

		// 根据扩展名推断语言
		switch mainExt {
		case ".go":
			info.Type = "Go 项目"
			info.MainLanguage = "Go"
		case ".js", ".ts", ".jsx", ".tsx":
			info.Type = "JavaScript/TypeScript 项目"
			info.MainLanguage = "JavaScript/TypeScript"
		case ".py":
			info.Type = "Python 项目"
			info.MainLanguage = "Python"
		case ".java":
			info.Type = "Java 项目"
			info.MainLanguage = "Java"
		case ".rs":
			info.Type = "Rust 项目"
			info.MainLanguage = "Rust"
		case ".php":
			info.Type = "PHP 项目"
			info.MainLanguage = "PHP"
		case ".rb":
			info.Type = "Ruby 项目"
			info.MainLanguage = "Ruby"
		default:
			info.Type = "通用项目"
			info.MainLanguage = "多种语言"
		}
	}

	// 收集所有检测到的语言
	for ext := range extCount {
		switch ext {
		case ".go":
			info.Languages = append(info.Languages, "Go")
		case ".js", ".ts", ".jsx", ".tsx":
			info.Languages = append(info.Languages, "JavaScript/TypeScript")
		case ".py":
			info.Languages = append(info.Languages, "Python")
		case ".java":
			info.Languages = append(info.Languages, "Java")
		case ".rs":
			info.Languages = append(info.Languages, "Rust")
		case ".php":
			info.Languages = append(info.Languages, "PHP")
		case ".rb":
			info.Languages = append(info.Languages, "Ruby")
		case ".html", ".htm":
			info.Languages = append(info.Languages, "HTML")
		case ".css", ".scss", ".less":
			info.Languages = append(info.Languages, "CSS")
		case ".md":
			info.Languages = append(info.Languages, "Markdown")
		case ".json", ".yaml", ".yml", ".toml":
			info.Languages = append(info.Languages, "配置文件")
		}
	}

	// 去重
	info.Languages = uniqueStrings(info.Languages)

	return nil
}

// analyzeDependencies 分析项目依赖
func analyzeDependencies(info *ProjectInfo) error {
	// 检查常见的依赖文件
	dependencyFiles := []string{
		"go.mod",
		"package.json",
		"Cargo.toml",
		"requirements.txt",
		"pyproject.toml",
		"pom.xml",
		"build.gradle",
		"composer.json",
		"Gemfile",
	}

	for _, depFile := range dependencyFiles {
		filePath := filepath.Join(info.Path, depFile)
		if _, err := os.Stat(filePath); err == nil {
			// 文件存在，记录依赖文件类型
			info.Dependencies[depFile] = []string{"已检测到依赖文件"}
		}
	}

	return nil
}

// analyzeDirectoryStructure 分析目录结构
func analyzeDirectoryStructure(info *ProjectInfo) error {
	// 遍历目录，收集主要目录结构
	err := filepath.Walk(info.Path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只显示顶层目录和重要文件
		relPath, _ := filepath.Rel(info.Path, path)
		if relPath == "." {
			return nil
		}

		// 只记录目录和重要文件
		if fi.IsDir() {
			// 跳过隐藏目录和常见忽略目录
			if strings.HasPrefix(fi.Name(), ".") ||
				fi.Name() == "node_modules" ||
				fi.Name() == "vendor" ||
				fi.Name() == "target" ||
				fi.Name() == "dist" ||
				fi.Name() == "build" {
				return filepath.SkipDir
			}

			// 只记录顶层目录
			if !strings.Contains(relPath, string(os.PathSeparator)) {
				info.Structure = append(info.Structure, relPath+"/")
			}
		} else {
			// 记录重要文件
			if !strings.Contains(relPath, string(os.PathSeparator)) {
				ext := strings.ToLower(filepath.Ext(fi.Name()))
				importantExts := []string{".go", ".js", ".ts", ".py", ".java", ".rs", ".php", ".rb", ".md"}
				for _, importantExt := range importantExts {
					if ext == importantExt {
						info.Structure = append(info.Structure, relPath)
						break
					}
				}
			}
		}

		return nil
	})

	return err
}

// generateProjectSummary 生成项目摘要
func generateProjectSummary(info *ProjectInfo) string {
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("项目名称: %s\n", info.Name))
	summary.WriteString(fmt.Sprintf("项目类型: %s\n", info.Type))
	summary.WriteString(fmt.Sprintf("主要语言: %s\n", info.MainLanguage))

	if len(info.Languages) > 0 {
		summary.WriteString(fmt.Sprintf("使用语言: %s\n", strings.Join(info.Languages, ", ")))
	}

	if len(info.Dependencies) > 0 {
		summary.WriteString("依赖管理: ")
		deps := []string{}
		for depFile := range info.Dependencies {
			deps = append(deps, depFile)
		}
		summary.WriteString(strings.Join(deps, ", "))
		summary.WriteString("\n")
	}

	if len(info.Structure) > 0 {
		summary.WriteString("目录结构:\n")
		for _, item := range info.Structure {
			summary.WriteString(fmt.Sprintf("  - %s\n", item))
		}
	}

	return summary.String()
}

// generateAgentMD 生成 AGENT.md 内容
func generateAgentMD(info *ProjectInfo) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s - 项目上下文文档\n\n", info.Name))
	content.WriteString(fmt.Sprintf("**生成时间**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	content.WriteString("## 项目概述\n\n")
	content.WriteString(fmt.Sprintf("- **项目名称**: %s\n", info.Name))
	content.WriteString(fmt.Sprintf("- **项目类型**: %s\n", info.Type))
	content.WriteString(fmt.Sprintf("- **主要语言**: %s\n", info.MainLanguage))

	if len(info.Languages) > 0 {
		content.WriteString(fmt.Sprintf("- **使用语言**: %s\n", strings.Join(info.Languages, ", ")))
	}

	content.WriteString("\n## 技术栈\n\n")

	if len(info.Dependencies) > 0 {
		content.WriteString("### 依赖管理\n")
		for depFile := range info.Dependencies {
			content.WriteString(fmt.Sprintf("- %s\n", depFile))
		}
		content.WriteString("\n")
	}

	content.WriteString("### 目录结构\n\n")
	if len(info.Structure) > 0 {
		for _, item := range info.Structure {
			content.WriteString(fmt.Sprintf("- %s\n", item))
		}
	} else {
		content.WriteString("（暂无目录结构信息）\n")
	}

	content.WriteString("\n## 开发约定\n\n")
	content.WriteString("### 代码风格\n")
	content.WriteString("- 遵循项目现有的编码规范和风格\n")
	content.WriteString("- 保持代码简洁、高效、可维护\n")
	content.WriteString("- 添加必要的注释和文档\n\n")

	content.WriteString("### 错误处理\n")
	content.WriteString("- 使用适当的错误处理机制\n")
	content.WriteString("- 错误信息要清晰明确\n")
	content.WriteString("- 记录重要的错误日志\n\n")

	content.WriteString("### 测试\n")
	content.WriteString("- 编写单元测试和集成测试\n")
	content.WriteString("- 确保测试覆盖核心功能\n")
	content.WriteString("- 定期运行测试套件\n\n")

	content.WriteString("## 注意事项\n\n")
	content.WriteString("1. 优先使用项目现有的库和框架\n")
	content.WriteString("2. 遵循项目的架构设计模式\n")
	content.WriteString("3. 保持向后兼容性\n")
	content.WriteString("4. 安全第一：不要引入安全漏洞\n")
	content.WriteString("5. 性能考虑：避免不必要的资源消耗\n\n")

	content.WriteString("---\n\n")
	content.WriteString("*本文档由 PolyAgent 自动生成，用于提供项目上下文信息*\n")

	return content.String()
}

// integrateAgentMDIntoSystemPrompt 将 AGENT.md 内容集成到系统提示词中
func integrateAgentMDIntoSystemPrompt() error {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}

	// 读取 AGENT.md 文件
	agentMDPath := filepath.Join(cwd, "AGENT.md")
	content, err := os.ReadFile(agentMDPath)
	if err != nil {
		return fmt.Errorf("读取 AGENT.md 失败: %v", err)
	}

	// 这里应该将 AGENT.md 内容添加到系统提示词中
	// 由于系统提示词是在 api_client.go 中动态生成的，
	// 我们需要修改系统提示词的生成逻辑来包含 AGENT.md 内容

	// 目前先记录到日志，后续可以扩展系统提示词生成逻辑
	fmt.Printf("AGENT.md 内容已准备集成到系统提示词中（%d 字节）\n", len(content))

	return nil
}

// uniqueStrings 去除字符串切片中的重复项
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, str := range input {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}
