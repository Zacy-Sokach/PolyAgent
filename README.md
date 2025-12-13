# PolyAgent

一个类似 Claude Code 的 Vibe Coding 工具，使用 Go 开发，基于 TUI 界面，直接调用 GLM-4.5 API。

## 功能特性

- 🎯 **Vibe Coding 工作流**：通过自然语言对话生成和修改代码
- 💬 **实时流式响应**：支持 GLM-4.5 的流式输出，实时显示生成内容
- 📁 **代码上下文感知**：自动读取当前工作目录的文件结构作为对话上下文
- 💾 **代码保存与插入**：一键将 AI 生成的代码保存到当前文件
- 🗂️ **历史会话管理**：自动保存对话历史
- 🔐 **安全的配置管理**：API Key 加密存储

## 安装

### 快速安装（推荐）

安装完成后，直接在终端输入 `polyagent` 即可运行。

#### Linux 和 macOS

```bash
curl -fsSL https://raw.githubusercontent.com/Zacy-Sokach/PolyAgent/main/scripts/install.sh | bash

# 安装完成后直接运行
polyagent
```

#### Windows

```powershell
irm https://raw.githubusercontent.com/Zacy-Sokach/PolyAgent/main/scripts/install.ps1 | iex

# 安装完成后直接运行
polyagent
```

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/Zacy-Sokach/PolyAgent.git
cd PolyAgent

# 编译
go build -o polyagent ./cmd/polyagent

# 运行
./polyagent
```

### 直接运行

```bash
go run ./cmd/polyagent
```

## 使用方法

1. **首次运行**：程序会提示输入 GLM-4.5 API Key
2. **基本操作**：
   - `Enter`：发送消息
   - `Ctrl+S`：将 AI 生成的代码保存到当前文件
   - `Esc`：取消正在进行的 AI 思考
   - `Ctrl+C`：退出程序（自动保存历史）

3. **Vibe Coding 工作流**：
   - 在代码目录中启动 PolyAgent
   - 用自然语言描述你想要的功能
   - AI 会基于当前目录上下文生成代码
   - 按 `Ctrl+S` 保存生成的代码到文件
   - 继续对话迭代改进

## 配置

配置文件位置：
- **Windows**: `%APPDATA%\polyagent\config.yaml` (例如: `C:\Users\用户名\AppData\Roaming\polyagent\config.yaml`)
- **Linux/macOS**: `~/.config/polyagent/config.yaml`

```yaml
api_key: your_glm_api_key
model: glm-4.5
```

## 项目结构

```
PolyAgent/
├── cmd/polyagent/          # 主程序入口
├── internal/
│   ├── api/               # GLM-4.5 API 客户端
│   ├── config/            # 配置管理
│   ├── tui/               # TUI 界面
│   └── utils/             # 工具函数
└── README.md
```

## 技术栈

- **语言**: Go 1.21+
- **TUI 框架**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **API**: GLM-4.5 (智谱AI)
- **配置**: YAML

## 开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建
go build ./cmd/polyagent
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件