# PolyAgent - iFlow 上下文文档

## 项目概述

**PolyAgent** 是一个类似 Claude Code 的 Vibe Coding 工具，使用 Go 语言开发，基于 TUI（终端用户界面），直接调用 GLM-4.5 API。该项目旨在提供一个沉浸式的 AI 辅助编程体验，让开发者通过自然语言对话来生成和修改代码。

### 核心概念：Vibe Coding
Vibe Coding（氛围编程）是一种新型编程范式，开发者通过自然语言描述需求，由 AI 生成代码，开发者只关注结果是否符合预期，不关心具体实现细节。

## 技术栈

- **编程语言**: Go 1.25.5+
- **TUI 框架**: Bubble Tea (charmbracelet/bubbletea)
- **UI 组件**: Bubbles (charmbracelet/bubbles)
- **样式**: Lip Gloss (charmbracelet/lipgloss)
- **配置管理**: YAML (gopkg.in/yaml.v3)
- **Markdown 渲染**: Goldmark (github.com/yuin/goldmark)
- **AI 模型**: GLM-4.5 (智谱AI)
- **工具集成**: Model Context Protocol (MCP) 支持

## 项目结构

```
PolyAgent/
├── cmd/polyagent/          # 主程序入口
│   └── main.go            # 程序启动和配置初始化
├── internal/              # 内部包，不对外暴露
│   ├── api/              # GLM-4.5 API 客户端
│   │   ├── client.go     # API 客户端实现（支持流式响应）
│   │   └── types.go      # API 数据结构定义
│   ├── config/           # 配置管理
│   │   ├── config.go     # 配置文件读写
│   │   └── config_test.go # 配置测试
│   ├── mcp/              # Model Context Protocol 工具集成
│   │   ├── handler.go    # 工具处理器
│   │   ├── protocol.go   # MCP 协议实现
│   │   ├── advanced_tools.go # 高级工具
│   │   ├── web_search_tool.go # 网络搜索工具
│   │   └── tool_test.go  # 工具测试
│   ├── tui/              # 终端用户界面
│   │   ├── model.go      # Bubble Tea 主模型
│   │   ├── messages.go   # 消息处理逻辑
│   │   ├── tool_integration.go # 工具集成管理
│   │   ├── command_parser.go   # 命令解析器
│   │   ├── editor_parser.go    # 编辑器集成
│   │   ├── init_command.go     # 初始化命令
│   │   ├── markdown.go         # Markdown 渲染器
│   │   ├── markdown_export.go  # Markdown 导出
│   │   └── model_helpers.go    # 模型辅助函数
│   └── utils/            # 工具函数
│       ├── context.go    # 代码上下文读取
│       ├── file_ops.go   # 文件操作
│       ├── editor.go     # 编辑器集成
│       ├── history.go    # 历史会话管理
│       └── context_test.go # 上下文测试
├── AGENT.md              # 项目上下文文档
├── CAPABILITIES.md       # 能力说明文档
├── IFLOW.md              # iFlow 上下文文档
├── go.mod                # Go 模块定义
├── go.sum                # 依赖校验
├── polyagent             # 编译后的可执行文件
├── LICENSE               # MIT 许可证
└── README.md             # 项目说明文档
```

## 构建和运行

### 安装依赖
```bash
go mod download
```

### 编译项目
```bash
# 编译为可执行文件
go build -o polyagent ./cmd/polyagent

# 或直接运行
go run ./cmd/polyagent
```

### 运行程序
```bash
./polyagent
```

**首次运行流程**：
1. 程序会检测是否配置了 GLM-4.5 API Key
2. 如果未配置，会提示用户输入 API Key
3. 配置保存在 `~/.config/polyagent/config.yaml`
4. 进入 TUI 交互界面

## 功能特性

### 1. Vibe Coding 工作流
- 通过自然语言对话生成和修改代码
- AI 基于当前工作目录的上下文生成代码
- 支持多轮对话迭代改进

### 2. 实时流式响应
- 支持 GLM-4.5 的流式输出
- 实时显示 AI 生成的内容
- 提供流畅的交互体验

### 3. 代码上下文感知
- 自动读取当前工作目录的文件结构
- 识别代码文件（Go, Python, JavaScript, Java 等）
- 将目录上下文作为提示词的一部分

### 4. 代码保存与插入
- 一键将 AI 生成的代码保存到当前文件
- 支持追加代码到文件末尾
- 快捷键：`Ctrl+S`

### 5. Markdown 渲染支持
- 内置 Markdown 渲染器，支持格式化显示
- 使用 Goldmark 引擎进行 Markdown 解析
- 在 TUI 中提供丰富的文本格式显示

### 6. MCP 工具集成
- 支持 Model Context Protocol (MCP)
- 集成多种工具，如文件操作、网络搜索等
- 安全的路径验证和权限控制

### 7. 历史会话管理
- 自动保存对话历史到 `~/.config/polyagent/history.json`
- 支持历史记录查看（需实现）
- 退出时自动保存

### 8. 安全的配置管理
- API Key 加密存储
- 配置文件使用 YAML 格式
- 配置路径：`~/.config/polyagent/config.yaml`

## TUI 交互快捷键

| 快捷键 | 功能 |
|--------|------|
| `Enter` | 发送消息 |
| `Ctrl+S` | 保存 AI 生成的代码到当前文件 |
| `Esc` | 取消正在进行的 AI 思考 |
| `Ctrl+C` | 退出程序（自动保存历史） |

## 开发约定

### 代码组织
- **internal/** 目录下的包不对外暴露，仅供内部使用
- **cmd/** 目录包含可执行程序的入口点
- 每个包有明确的单一职责

### 错误处理
- 使用 Go 的错误处理模式
- 错误信息要清晰明确
- 配置文件和 API 调用错误需要友好提示

### 配置管理
- 配置文件使用 YAML 格式
- 配置路径：`~/.config/polyagent/`
- 支持默认配置和用户自定义配置

## API 集成

### GLM-4.5 API 端点
- 基础 URL: `https://open.bigmodel.cn/api/paas/v4`
- 聊天补全端点: `/chat/completions`
- 支持流式响应（Server-Sent Events）

### 请求参数
```json
{
  "model": "glm-4.5",
  "messages": [...],
  "stream": true,
  "max_tokens": 4096,
  "temperature": 0.6,
  "thinking": {
    "type": "enabled"
  },
  "tools": [...]
}
```

## MCP 工具系统

### 支持的工具类型
- **文件操作工具**: 安全的文件读写、目录遍历
- **网络搜索工具**: 集成网络搜索功能
- **代码分析工具**: 代码理解和分析
- **系统工具**: 环境信息获取

### 工具安全机制
- 路径验证：限制访问安全目录范围
- 权限控制：细粒度的操作权限管理
- 错误处理：完善的错误捕获和恢复机制

## 已实现功能

1. **MCP 工具系统**：完整的 Model Context Protocol 实现
2. **Markdown 渲染**：内置 Markdown 渲染器支持
3. **工具集成管理**：统一的工具注册和调用机制
4. **命令解析系统**：增强的命令解析和处理
5. **编辑器集成**：与外部编辑器的集成支持
6. **安全路径验证**：完善的文件访问安全机制

## 待实现功能

1. **历史会话查看**：实现历史记录的查看和加载功能
2. **多文件支持**：支持在多个文件间切换和操作
3. **代码编辑**：集成基本的代码编辑功能
4. **测试套件完善**：扩展单元测试和集成测试覆盖
5. **配置管理界面**：在 TUI 中管理配置
6. **插件系统**：支持第三方工具插件扩展

## 故障排除

### 常见问题

1. **API Key 错误**
   - 检查 `~/.config/polyagent/config.yaml` 文件
   - 确认 API Key 有效且未过期
   - 重新运行程序输入新的 API Key

2. **编译错误**
   - 运行 `go mod tidy` 整理依赖
   - 确保 Go 版本 >= 1.25.5
   - 检查依赖包是否完整

3. **TUI 显示问题**
   - 确保终端支持 ANSI 转义序列
   - 检查终端尺寸是否足够
   - 尝试调整终端字体

### 调试模式
目前没有内置调试模式，可以通过查看日志文件：
- 配置日志：`~/.config/polyagent/config.yaml`
- 历史日志：`~/.config/polyagent/history.json`

## 贡献指南

1. Fork 项目仓库
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

### 代码风格
- 遵循 Go 官方代码风格
- 使用 `gofmt` 格式化代码
- 添加必要的注释

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

*本文档最后更新：2025-12-07*  
*项目状态：功能持续增强，MCP 工具系统已集成*