# Monitor

一个轻量级的 Go 语言服务器监控系统，类似于哪吒监控，提供简单易用的服务端、前端管理界面和跨平台 Agent。

## 功能特性

- 🔐 **安全认证**: Admin 用户登录系统，JWT 认证
- 📊 **指标收集**: CPU、内存、磁盘、网络、负载等基础指标
- 📈 **历史数据**: SQLite 存储历史数据，React+MUI 可视化展示
- 🚨 **告警系统**: 支持阈值触发的告警规则
- 📧 **邮件通知**: 告警邮件通知功能
- 💻 **跨平台 Agent**: 支持 Linux/Windows/macOS
- ⚡ **实时监控**: 秒级数据上报，心跳检测
- 🔒 **安全传输**: TLS + Brotli 压缩
- 📝 **JSON 通信**: 简单的 JSON 格式数据交换
- 🚀 **低开销**: 性能开销小，易于部署

## 快速开始

### 安装依赖

确保已安装 Go 1.21+ 和 Node.js 16+。

```bash
# 克隆仓库
git clone https://github.com/jyxjjj/Monitor.git
cd Monitor

# 安装 Go 依赖
go mod download

# 构建前端
cd frontend
npm install
npm run build
cd ..
```

### 构建

```bash
# 构建服务器
go build -o monitor-server ./cmd/server

# 构建 Agent
go build -o monitor-agent ./cmd/agent
```

### 配置

#### 服务器配置 (server-config.json)

首次运行服务器会自动生成默认配置文件：

```json
{
  "server_addr": ":8443",
  "tls_cert_file": "",
  "tls_key_file": "",
  "db_path": "./monitor.db",
  "admin_password": "admin123",
  "smtp_host": "",
  "smtp_port": 587,
  "smtp_user": "",
  "smtp_password": "",
  "email_from": "",
  "alert_email": ""
}
```

**注意**: 请修改 `admin_password` 为安全密码。如需启用 TLS，配置证书路径。

#### Agent 配置 (agent-config.json)

首次运行 Agent 会自动生成默认配置文件：

```json
{
  "server_url": "https://localhost:8443",
  "agent_id": "hostname",
  "agent_name": "hostname",
  "report_interval": 5,
  "tls_skip_verify": true
}
```

### 运行

#### 启动服务器

```bash
./monitor-server
```

服务器默认在 `:8443` 端口启动（或配置文件中指定的端口）。

访问 `http://localhost:8443` 打开 Web 管理界面。
默认密码: `admin123`

#### 启动 Agent

```bash
./monitor-agent
```

Agent 会自动连接到服务器并开始上报监控数据。

## 生成 TLS 证书（可选）

为了启用 HTTPS，可以生成自签名证书：

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"
```

然后在 `server-config.json` 中配置：

```json
{
  "tls_cert_file": "server.crt",
  "tls_key_file": "server.key"
}
```

## 告警配置

在 Web 界面中配置告警规则：

1. 登录后点击 "Alert Rules"
2. 点击 "Add Rule"
3. 配置告警条件（指标类型、阈值、持续时间等）
4. 启用规则

告警类型：
- CPU 使用率
- 内存使用率
- 磁盘使用率
- 系统负载

## 邮件通知

在 `server-config.json` 中配置 SMTP 信息：

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_user": "your-email@gmail.com",
  "smtp_password": "your-app-password",
  "email_from": "your-email@gmail.com",
  "alert_email": "alert-recipient@example.com"
}
```

## 项目结构

```
Monitor/
├── cmd/
│   ├── server/      # 服务器入口
│   └── agent/       # Agent 入口
├── pkg/
│   ├── models/      # 数据模型
│   ├── collector/   # 指标收集器
│   ├── server/      # 服务器逻辑
│   ├── agent/       # Agent 逻辑
│   ├── compress/    # Brotli 压缩
│   └── config/      # 配置管理
└── frontend/        # React 前端
```

## API 端点

- `POST /api/login` - 管理员登录
- `GET /api/agents` - 获取 Agent 列表
- `GET /api/metrics/{agentId}` - 获取指标历史
- `POST /api/metrics/report` - Agent 上报数据
- `GET /api/alerts` - 获取告警列表
- `GET /api/alert-rules` - 获取告警规则
- `POST /api/alert-rules` - 创建告警规则
- `GET /api/config` - 获取配置信息

## 系统要求

### 服务器
- Go 1.21+
- 512MB RAM（最小）
- 10GB 磁盘空间

### Agent
- 支持的操作系统：Linux, Windows, macOS
- 最小 CPU 和内存占用

## 开发

### 运行前端开发服务器

```bash
cd frontend
npm start
```

前端开发服务器会在 `http://localhost:3000` 启动。

### 构建前端

```bash
cd frontend
npm run build
```

构建后的文件在 `frontend/build` 目录。

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- 灵感来源于 [哪吒监控](https://github.com/naiba/nezha)
- 使用 [gopsutil](https://github.com/shirou/gopsutil) 进行系统指标收集
- UI 基于 [Material-UI](https://mui.com/)