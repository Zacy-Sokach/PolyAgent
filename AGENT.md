# PolyAgent - 项目上下文文档

**生成时间**: 2025-12-07 18:06:00

## 项目概述

- **项目名称**: PolyAgent
- **项目类型**: Go 项目
- **主要语言**: Go 1.25.5+
- **使用框架**: TUI (Terminal User Interface)
- **项目描述**: 一个类似 Claude Code 的 Vibe Coding 工具，使用 Go 开发，基于 TUI 界面，直接调用 GLM-4.5 API

### 核心概念：Vibe Coding
Vibe Coding（氛围编程）是一种新型编程范式，开发者通过自然语言描述需求，由 AI 生成代码，开发者只关注结果是否符合预期，不关心具体实现细节。这种工作流大大提高了开发效率，特别适合快速原型开发和创意编程。

### 项目目标
- 提供类似 Claude Code 的 AI 辅助编程体验
- 通过自然语言对话实现代码生成和修改
- 集成丰富的工具生态系统，支持文件操作、网络搜索等功能
- 构建安全、稳定、可扩展的 AI 编程助手

## 技术栈和依赖

### 核心技术栈
- **编程语言**: Go 1.25.5+
- **TUI 框架**: Bubble Tea (charmbracelet/bubbletea v1.3.10)
- **UI 组件**: Bubbles (charmbracelet/bubbles v0.21.0)
- **样式引擎**: Lip Gloss (charmbracelet/lipgloss v1.1.0)
- **配置管理**: YAML (gopkg.in/yaml.v3)
- **Markdown 渲染**: Goldmark (github.com/yuin/goldmark v1.7.13)

### AI 集成
- **AI 模型**: GLM-4.5 (智谱AI)
- **API 端点**: `https://open.bigmodel.cn/api/paas/v4`
- **流式响应**: 支持 Server-Sent Events (SSE)
- **工具调用**: 支持 Function Calling 和 Model Context Protocol (MCP)

### 依赖关系
```go
module github.com/Zacy-Sokach/PolyAgent

go 1.25.5

require (
    github.com/charmbracelet/bubbles v0.21.0
    github.com/charmbracelet/bubbletea v1.3.10
    github.com/charmbracelet/lipgloss v1.1.0
    github.com/yuin/goldmark v1.7.13
    gopkg.in/yaml.v3 v3.0.1
)
```

### 间接依赖
- **终端处理**: charmbracelet/x/ansi, charmbracelet/x/term
- **颜色支持**: charmbracelet/colorprofile, lucasb-eyer/go-colorful
- **文本处理**: rivo/uniseg, mattn/go-runewidth
- **系统调用**: golang.org/x/sys, golang.org/x/text

## 项目结构说明

### 目录结构
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
│   │   ├── model_helpers.go    # 模型辅助函数
│   │   └── tools_prompt_generator.go # 工具提示生成器
│   └── utils/            # 工具函数
│       ├── context.go    # 代码上下文读取
│       ├── context_test.go # 上下文测试
│       ├── editor.go     # 编辑器集成
│       └── history.go    # 历史会话管理
├── AGENT.md              # 项目上下文文档
├── CAPABILITIES.md       # 能力说明文档
├── IFLOW.md              # iFlow 上下文文档
├── go.mod                # Go 模块定义
├── go.sum                # 依赖校验
├── polyagent             # 编译后的可执行文件
├── LICENSE               # Apache License 2.0
└── README.md             # 项目说明文档
```

### 核心模块说明

#### 1. API 模块 (`internal/api/`)
- **client.go**: GLM-4.5 API 客户端实现
  - 支持流式和非流式请求
  - 工具调用支持
  - 错误处理和重试机制
- **types.go**: API 数据结构定义
  - 消息格式定义
  - 工具调用结构
  - 响应解析结构

#### 2. 配置模块 (`internal/config/`)
- **config.go**: 配置文件管理
  - YAML 格式配置文件
  - API Key 安全存储
  - 默认配置管理
  - 配置路径：`~/.config/polyagent/config.yaml`

#### 3. MCP 模块 (`internal/mcp/`)
- **handler.go**: 工具处理器
- **protocol.go**: MCP 协议实现
- **advanced_tools.go**: 高级工具实现
- **web_search_tool.go**: 网络搜索工具
- **tool_test.go**: 工具测试

#### 4. TUI 模块 (`internal/tui/`)
- **model.go**: Bubble Tea 主模型
  - 状态管理
  - 消息处理
  - 工具集成
- **messages.go**: 消息处理逻辑
- **tool_integration.go**: 工具集成管理
- **command_parser.go**: 命令解析器
- **editor_parser.go**: 编辑器集成
- **init_command.go**: 初始化命令
- **markdown.go**: Markdown 渲染器
- **markdown_export.go**: Markdown 导出
- **model_helpers.go**: 模型辅助函数
- **tools_prompt_generator.go**: 工具提示生成器

#### 5. 工具模块 (`internal/utils/`)
- **context.go**: 代码上下文读取
  - 目录结构分析
  - 文件内容读取
  - 代码文件识别
- **context_test.go**: 上下文测试
- **editor.go**: 编辑器集成
- **history.go**: 历史会话管理

## 核心功能

### 1. Vibe Coding 工作流
- **自然语言编程**: 通过对话生成代码
- **上下文感知**: 自动分析当前项目结构
- **迭代改进**: 支持多轮对话优化代码
- **代码保存**: 一键保存生成的代码

### 2. 实时流式响应
- **SSE 支持**: 基于 Server-Sent Events 的流式输出
- **实时显示**: 逐字显示 AI 生成内容
- **思考过程**: 显示 AI 的思考过程（reasoning）
- **中断控制**: 可随时中断 AI 思考

### 3. 代码上下文感知
- **目录分析**: 自动读取当前工作目录结构
- **文件识别**: 智能识别代码文件类型
- **内容读取**: 读取关键文件内容作为上下文
- **深度限制**: 限制目录遍历深度（最多5层）

### 4. MCP 工具集成
- **文件操作**: 安全的文件读写、目录遍历
- **网络搜索**: 集成 Tavily API 进行网络搜索
- **代码分析**: 代码理解和分析工具
- **系统工具**: 环境信息获取、Git 操作等

### 5. Markdown 渲染支持
- **实时渲染**: 在 TUI 中实时渲染 Markdown
- **语法高亮**: 代码块语法高亮
- **格式支持**: 支持标准 Markdown 语法
- **导出功能**: 支持导出为 Markdown 文件

### 6. 历史会话管理
- **自动保存**: 自动保存对话历史
- **持久化存储**: 历史记录持久化到文件
- **加载恢复**: 支持加载历史会话
- **隐私保护**: 敏感信息过滤

## 开发约定和最佳实践

### 代码组织原则
1. **模块化设计**: 每个包有明确的单一职责
2. **内部封装**: `internal/` 目录下的包不对外暴露
3. **清晰分层**: API 层、业务逻辑层、UI 层分离
4. **依赖注入**: 通过参数传递依赖，避免全局变量

### 错误处理约定
1. **错误返回**: 使用 Go 的标准错误处理模式
2. **错误信息**: 错误信息要清晰明确，包含上下文
3. **错误分类**: 区分用户错误、系统错误、网络错误
4. **错误恢复**: 关键操作要有错误恢复机制

### 配置管理规范
1. **配置格式**: 统一使用 YAML 格式
2. **配置路径**: `~/.config/polyagent/`
3. **敏感信息**: API Key 等敏感信息加密存储
4. **默认配置**: 提供合理的默认配置

### TUI 开发规范
1. **响应式设计**: 适配不同终端尺寸
2. **用户体验**: 提供清晰的视觉反馈
3. **快捷键**: 统一的快捷键设计
4. **性能优化**: 优化长消息渲染，避免界面卡顿

### API 集成规范
1. **流式处理**: 优先使用流式响应
2. **工具调用**: 标准化的工具调用格式
3. **错误处理**: 完善的网络错误处理
4. **超时控制**: 合理的请求超时设置

### 安全规范
1. **路径验证**: 严格的文件路径验证
2. **权限控制**: 细粒度的操作权限管理
3. **输入验证**: 所有用户输入都要验证
4. **敏感信息**: API Key 等敏感信息不记录日志

## 构建和运行指南

### 环境要求
- Go 1.25.5 或更高版本
- 支持 ANSI 转义序列的终端
- 有效的 GLM-4.5 API Key

### 安装步骤

#### 1. 获取源码
```bash
git clone https://github.com/Zacy-Sokach/PolyAgent.git
cd PolyAgent
```

#### 2. 安装依赖
```bash
go mod download
go mod tidy
```

#### 3. 编译项目
```bash
# 编译为可执行文件
go build -o polyagent ./cmd/polyagent

# 或直接运行
go run ./cmd/polyagent
```

#### 4. 运行程序
```bash
./polyagent
```

### 首次运行配置
1. **API Key 配置**: 程序会检测是否配置了 GLM-4.5 API Key
2. **输入 API Key**: 如果未配置，会提示用户输入 API Key
3. **Tavily API Key**: 可选配置，用于网络搜索功能
4. **配置保存**: 配置保存在 `~/.config/polyagent/config.yaml`

### 配置文件格式
```yaml
api_key: "your_glm_api_key"
model: "glm-4.5"
tavily_api_key: "your_tavily_api_key"  # 可选
```

### 开发环境设置

#### 1. 开发工具
```bash
# 安装开发工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### 2. 运行测试
```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行特定包的测试
go test ./internal/config
```

#### 3. 代码格式化
```bash
# 格式化代码
go fmt ./...

# 导入排序
goimports -w .
```

#### 4. 代码检查
```bash
# 运行 golangci-lint
golangci-lint run
```

### 构建选项

#### 1. 交叉编译
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o polyagent-linux-amd64 ./cmd/polyagent

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o polyagent-darwin-amd64 ./cmd/polyagent

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o polyagent-windows-amd64.exe ./cmd/polyagent
```

#### 2. 优化构建
```bash
# 优化构建（减小体积）
go build -ldflags="-s -w" -o polyagent ./cmd/polyagent

# 进一步压缩（需要 upx 工具）
upx --best polyagent
```

## TUI 交互指南

### 快捷键说明

| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `Enter` | 发送消息 | 发送用户输入到 AI |
| `Ctrl+S` | 保存代码 | 将 AI 生成的代码保存到当前文件 |
| `Ctrl+R` | 回退会话 | 回退当前会话的所有修改 |
| `Esc` | 取消思考 | 中断正在进行的 AI 思考过程 |
| `Ctrl+C` | 退出程序 | 退出程序并自动保存历史 |

### 界面元素说明

#### 1. 消息显示区域
- **用户消息**: 蓝色显示，前缀 "你: "
- **AI 消息**: 绿色显示，前缀 "AI: "
- **系统消息**: 紫色显示，前缀 "系统: "
- **思考过程**: 特殊格式显示 AI 的思考过程

#### 2. 输入区域
- **多行输入**: 支持多行文本输入
- **自动换行**: 自动处理文本换行
- **历史记录**: 支持上下箭头浏览输入历史

#### 3. 状态指示
- **思考状态**: 显示 "AI正在思考中..."
- **工具调用**: 显示工具调用信息
- **错误提示**: 显示错误和警告信息

### 命令系统

#### 1. 内置命令
- **`/init`**: 初始化项目分析
- **`/help`**: 显示帮助信息
- **`/clear`**: 清除当前会话
- **`/save`**: 保存当前会话

#### 2. 命令格式
```
/命令名 [参数]
```

#### 3. 自定义命令
支持通过插件系统扩展自定义命令

## 注意事项

### 安全注意事项

#### 1. API Key 安全
- **存储位置**: API Key 存储在 `~/.config/polyagent/config.yaml`
- **权限设置**: 确保配置文件权限正确 (600)
- **不分享**: 不要将 API Key 提交到版本控制系统
- **定期更换**: 建议定期更换 API Key

#### 2. 文件操作安全
- **路径验证**: 严格验证文件路径，防止路径遍历攻击
- **权限检查**: 检查文件读写权限
- **备份重要**: 操作重要文件前先备份
- **限制范围**: 限制文件操作范围在项目目录内

#### 3. 网络安全
- **HTTPS**: 所有网络请求使用 HTTPS
- **超时设置**: 合理的网络请求超时设置
- **输入验证**: 验证所有网络输入
- **错误处理**: 完善的网络错误处理

### 性能优化建议

#### 1. 内存使用
- **消息限制**: 限制历史消息数量（默认50条）
- **渲染优化**: 优化长消息渲染，避免卡顿
- **缓存机制**: 使用缓存减少重复计算
- **垃圾回收**: 及时清理不再使用的资源

#### 2. 网络优化
- **连接复用**: 复用 HTTP 连接
- **请求合并**: 合并多个小请求
- **缓存响应**: 缓存重复的网络请求
- **错误重试**: 合理的错误重试机制

#### 3. 用户体验
- **响应速度**: 优化界面响应速度
- **视觉反馈**: 提供清晰的操作反馈
- **错误提示**: 友好的错误提示信息
- **帮助文档**: 完善的帮助文档

### 常见问题解决

#### 1. API Key 相关问题
**问题**: API Key 无效或过期
**解决**: 
- 检查 API Key 是否正确
- 确认 API Key 未过期
- 重新配置 API Key

**问题**: API 调用失败
**解决**:
- 检查网络连接
- 确认 API 服务状态
- 查看错误日志

#### 2. 编译问题
**问题**: Go 版本不兼容
**解决**:
- 升级 Go 到 1.25.5+
- 更新依赖包
- 检查 go.mod 文件

**问题**: 依赖包缺失
**解决**:
- 运行 `go mod tidy`
- 检查网络连接
- 清理模块缓存：`go clean -modcache`

#### 3. TUI 显示问题
**问题**: 终端显示异常
**解决**:
- 确认终端支持 ANSI 转义序列
- 调整终端字体和大小
- 尝试不同的终端模拟器

**问题**: 界面卡顿
**解决**:
- 减少历史消息数量
- 优化长消息渲染
- 检查系统资源使用情况

#### 4. 配置问题
**问题**: 配置文件读取失败
**解决**:
- 检查配置文件路径
- 确认文件权限
- 验证 YAML 格式

**问题**: API Key 配置错误
**解决**:
- 重新运行程序配置 API Key
- 手动编辑配置文件
- 检查 API Key 格式

## 许可证

本项目采用 Apache License 2.0 许可证。详见 [LICENSE](LICENSE) 文件。

### 许可证要点
- **商业使用**: 允许商业使用
- **修改**: 允许修改和分发
- **专利**: 明确的专利授权
- **责任**: 免责声明
- **商标**: 商标使用限制

## 贡献指南

### 如何贡献

#### 1. Fork 项目
```bash
# Fork 并克隆项目
git clone https://github.com/YOUR_USERNAME/PolyAgent.git
cd PolyAgent
```

#### 2. 创建功能分支
```bash
# 创建新分支
git checkout -b feature/your-feature-name
```

#### 3. 开发和测试
```bash
# 运行测试
go test ./...

# 运行代码检查
golangci-lint run

# 格式化代码
go fmt ./...
```

#### 4. 提交更改
```bash
# 添加更改
git add .

# 提交更改
git commit -m "feat: add your feature description"
```

#### 5. 推送和创建 PR
```bash
# 推送分支
git push origin feature/your-feature-name

# 在 GitHub 创建 Pull Request
```

### 代码风格指南

#### 1. Go 代码风格
- 遵循 [Go 官方代码风格](https://golang.org/doc/effective_go.html)
- 使用 `go fmt` 格式化代码
- 函数和变量使用驼峰命名法
- 包名使用小写字母

#### 2. 注释规范
- 为导出的类型、函数、常量添加注释
- 注释应该解释"为什么"而不是"是什么"
- 使用完整的句子，以句号结尾

```go
// User represents a user in the system
type User struct {
    ID       string    // Unique identifier for the user
    Name     string    // User's display name
    Email    string    // User's email address
    Created  time.Time // When the user was created
}
```

#### 3. 错误处理
- 使用 `errors.New()` 或 `fmt.Errorf()` 创建错误
- 错误信息应该以小写字母开头
- 包含足够的上下文信息

```go
// 好的错误处理
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// 不好的错误处理
if err != nil {
    return err
}
```

#### 4. 测试规范
- 为每个包编写测试
- 测试文件命名为 `*_test.go`
- 使用表驱动测试

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -1, -2},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型说明
- **feat**: 新功能
- **fix**: 修复 bug
- **docs**: 文档更改
- **style**: 代码格式化（不影响代码运行的变动）
- **refactor**: 重构（既不是新增功能，也不是修改 bug 的代码变动）
- **perf**: 性能优化
- **test**: 增加测试
- **chore**: 构建过程或辅助工具的变动

#### 示例
```
feat(api): add streaming support for GLM-4.5

Add streaming response support using Server-Sent Events.
This allows real-time display of AI-generated content.

Closes #123
```

### 问题报告

#### 1. Bug 报告
使用 GitHub Issues 报告 bug，请包含：
- 问题描述
- 重现步骤
- 期望行为
- 实际行为
- 环境信息（Go 版本、操作系统等）
- 相关日志或错误信息

#### 2. 功能请求
提出新功能建议时，请说明：
- 功能描述
- 使用场景
- 期望的行为
- 可能的实现方案

## 版本管理

### 版本号规则
使用 [Semantic Versioning](https://semver.org/)：
- **主版本号**: 不兼容的 API 修改
- **次版本号**: 向下兼容的功能性新增
- **修订号**: 向下兼容的问题修正

### 发布流程
1. 更新版本号
2. 更新 CHANGELOG.md
3. 创建 Git tag
4. 创建 GitHub Release

## 联系方式

- **项目地址**: https://github.com/Zacy-Sokach/PolyAgent
- **问题反馈**: https://github.com/Zacy-Sokach/PolyAgent/issues
- **讨论**: https://github.com/Zacy-Sokach/PolyAgent/discussions

## 致谢

感谢以下开源项目和贡献者：

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 强大的 TUI 框架
- [GLM-4.5](https://open.bigmodel.cn/) - 智谱AI 的大语言模型
- [Go 社区](https://golang.org/) - 优秀的编程语言和工具链

---

*本文档由 PolyAgent 自动生成，用于提供项目上下文信息*  
*最后更新：2025-12-07 18:06:00*