package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Zacy-Sokach/PolyAgent/internal/config"
	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
	"github.com/Zacy-Sokach/PolyAgent/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	Version = "dev"
)

func main() {
	// 处理命令行参数
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("PolyAgent %s\n", Version)
			os.Exit(0)
		case "-h", "--help":
			fmt.Println("PolyAgent - Vibe Coding Tool")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  polyagent              Start the interactive TUI")
			fmt.Println("  polyagent -v, --version  Show version information")
			fmt.Println("  polyagent -h, --help     Show help information")
			fmt.Println()
			fmt.Println("Commands in TUI:")
			fmt.Println("  check update           Check for updates")
			fmt.Println("  update                 Update PolyAgent to latest version")
			fmt.Println("  /init                  Initialize project documentation")
			os.Exit(0)
		}
	}
	
	// 添加panic恢复
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("程序发生panic: %v\n", r)
			fmt.Println("堆栈跟踪:")
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if cfg.APIKey == "" {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("欢迎使用 PolyAgent!"))
		fmt.Println("首次使用需要配置 GLM-4.5 API Key")
		fmt.Print("请输入你的 GLM API Key: ")

		var apiKey string
		fmt.Scanln(&apiKey)

		cfg.APIKey = apiKey
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("GLM API Key 已保存!"))
	}

	// 检查 Tavily API Key（用于搜索功能）
	if cfg.TavilyAPIKey == "" {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("💡 检测到未配置 Tavily API Key"))
		fmt.Println("Tavily API Key 用于网页搜索和爬取功能 (web_search, web_crawl)")
		fmt.Println("如果暂时不需要使用搜索功能，可以直接回车跳过")
		fmt.Println()
		fmt.Println("获取免费 API Key: https://tavily.com/")
		fmt.Print("请输入 Tavily API Key（直接回车跳过）: ")

		var tavilyKey string
		fmt.Scanln(&tavilyKey)

		if tavilyKey != "" {
			cfg.TavilyAPIKey = tavilyKey
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("保存配置失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ Tavily API Key 已保存!"))
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("跳过配置，搜索功能将在首次使用时提示配置"))
		}
	}

	// 检查是否在交互式终端中
	if isTerminal() {
		// 创建 ToolRegistry，传入 FileEngine 配置（转换类型）
		fileEngineConfig := mcp.FileEngineConfig{
			AllowedRoots:    cfg.FileEngine.AllowedRoots,
			BlacklistedExts: cfg.FileEngine.BlacklistedExts,
			MaxFileSize:     cfg.FileEngine.MaxFileSize,
			EnableCache:     cfg.FileEngine.EnableCache,
			BackupDir:       cfg.FileEngine.BackupDir,
		}
		toolRegistry := mcp.DefaultToolRegistry(&fileEngineConfig)
		toolManager := tui.NewToolManagerWithRegistry(toolRegistry)
		
		// 暂时注释掉版本设置
		// tui.Version = Version
		
		// 创建模型并使用指针
		model := tui.InitialModel(cfg.APIKey, toolManager)
		p := tea.NewProgram(&model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("程序运行错误: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 非交互式环境，使用简单模式
		fmt.Println("PolyAgent 运行在非交互式模式")
		fmt.Println("请确保在交互式终端中运行以获得完整TUI体验")
		fmt.Printf("当前API Key: %s\n", maskAPIKey(cfg.APIKey))
		fmt.Println("程序将在非交互式环境中退出")
	}
}

func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}
