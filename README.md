# Coder AI 助手

基于Golang开发的智能AI助手系统，使用Eino框架和OpenAI/DeepSeek模型，支持MCP工具调用。

## 系统架构

该项目采用模块化设计，主要组件包括：

```
coder/
├── api/                # API接口定义
├── app/                # 应用程序初始化和上下文
├── config/             # 配置文件加载
├── internal/           # 内部包
│   ├── agent/          # AI代理实现
│   ├── cache/          # 缓存实现
│   ├── config/         # 配置处理
│   ├── handler/        # HTTP处理器
│   ├── mcp/            # MCP工具集成
│   ├── models/         # 数据模型
│   ├── server/         # HTTP服务器
│   └── tools/          # 工具函数
├── nodelog/            # 日志管理
├── static/             # 静态资源文件
├── config.toml         # 配置文件
├── go.mod              # Go模块文件
├── go.sum              # 依赖校验文件
└── main.go             # 主入口文件
```

## 技术栈

- Golang 1.23
- Gin Web框架
- Eino AI框架
- DeepSeek/OpenAI语言模型集成
- MCP工具链接器

## Windows开发环境配置

### 1. 安装Golang

1. 前往[Golang官方网站](https://golang.org/dl/)下载Windows安装包
2. 运行安装包，按照向导完成安装
   - 默认安装路径通常为 `C:\Program Files\Go` 或 `C:\Go`
   - 安装程序会自动设置PATH环境变量
3. 验证安装：打开命令提示符或PowerShell，输入 `go version`

### 2. 设置工作目录

```powershell
# 创建工作目录(以D盘为例)
mkdir D:\Projects\coder
cd D:\Projects\coder

# 克隆仓库（如适用）
git clone <repository-url> .
```

### 3. 安装项目依赖

```powershell
go mod download
go mod tidy
```

## 配置项目

1. 复制示例配置文件

```powershell
copy config.example.toml config.toml
```

2. 编辑配置文件，设置API密钥和其他必要配置：

```toml
# 重要: 设置你的OpenAI/DeepSeek API密钥
[openai]
api_key = "your-api-key"
base_url = "https://api.deepseek.com/api/v1"  # 可选，使用不同的API基础URL
```

## 运行项目

### 开发模式

```powershell
go run main.go
```

### 构建和部署

```powershell
# 构建可执行文件
go build -o coder-app.exe

# 运行应用
.\coder-app.exe
```

## API文档

主要API端点:

- `POST /v1/chat` - 发送聊天请求
- `GET /healthz` - 健康检查
- `GET /` - 静态前端资源

## MCP工具配置

系统支持MCP(Multi-Call Protocol)工具调用，在config.toml中配置：

```toml
[mcp]
enabled = true

[[mcp.clients]]
name = "工具名称"
enabled = true
url = "http://tool-service-url/sse"
description = "工具描述"
```

## 常见问题排查

1. **找不到Go命令**
   - 确保Go已添加到PATH环境变量中
   - 可以手动添加：右键"此电脑" > 属性 > 高级系统设置 > 环境变量 > 在Path中添加Go安装目录的bin文件夹

2. **依赖下载失败**
   - 尝试设置GOPROXY：`go env -w GOPROXY=https://goproxy.cn,direct`
   - 确保防火墙或杀毒软件未阻止Go访问网络

3. **端口被占用**
   - 在配置文件中修改端口号
   - 使用`netstat -ano | findstr "8080"`查看端口占用情况

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开Pull Request

## 许可证

[MIT License](LICENSE)

## 联系方式

如有问题，请联系项目维护者。
