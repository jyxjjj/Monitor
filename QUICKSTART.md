# 快速开始指南

5分钟快速部署 Monitor 监控系统。

## 步骤 1: 下载和构建

```bash
# 克隆仓库
git clone https://github.com/jyxjjj/Monitor.git
cd Monitor

# 安装依赖并构建
go mod download
make build
```

## 步骤 2: 启动服务器

```bash
# 首次运行生成默认配置
./monitor-server

# 编辑配置（可选，修改密码等）
vim server-config.json

# 再次启动服务器
./monitor-server
```

服务器将在 `:8443` 端口启动（或配置文件中指定的端口）。

## 步骤 3: 访问 Web 界面

打开浏览器访问：
- HTTP: http://localhost:8443
- HTTPS: https://localhost:8443 (如果配置了 TLS)

默认登录密码: `admin123`

**重要**: 请立即修改密码！

## 步骤 4: 启动 Agent

在需要监控的服务器上：

```bash
# 首次运行生成默认配置
./monitor-agent

# 编辑配置，修改服务器地址
vim agent-config.json
```

修改 `agent-config.json` 中的 `server_url` 为你的服务器地址：

```json
{
  "server_url": "http://your-server-ip:8443",
  "agent_id": "server-01",
  "agent_name": "My Server 01",
  "report_interval": 5,
  "tls_skip_verify": false
}
```

启动 Agent：

```bash
./monitor-agent
```

## 步骤 5: 配置告警（可选）

在 Web 界面中：

1. 点击 "Alert Rules"
2. 点击 "Add Rule"
3. 配置告警条件：
   - Agent: 选择要监控的 Agent
   - Metric: 选择监控指标（CPU/内存/磁盘/负载）
   - Condition: 设置阈值
   - Duration: 持续时间
   - Description: 告警描述

4. 启用规则

## 步骤 6: 配置邮件通知（可选）

编辑 `server-config.json`：

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_user": "your-email@gmail.com",
  "smtp_password": "your-app-password",
  "email_from": "your-email@gmail.com",
  "alert_email": "admin@example.com"
}
```

重启服务器使配置生效。

## 快速测试

### 测试 API

```bash
# 登录获取 token
curl -X POST http://localhost:8443/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'

# 查看 Agent 列表
curl http://localhost:8443/api/agents \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 下一步

- 阅读完整 [README.md](README.md) 了解更多功能
- 查看 [DEPLOYMENT.md](DEPLOYMENT.md) 了解生产环境部署
- 浏览 [API 文档](API.md) 了解 API 使用

## 常见问题

### Agent 显示离线？

- 检查 Agent 是否正常运行
- 确认 `server_url` 配置正确
- 检查网络连接和防火墙

### 无法登录？

- 确认密码正确（默认 `admin123`）
- 检查浏览器控制台是否有错误

### 没有数据显示？

- 确认 Agent 已启动并成功连接
- 等待几秒让 Agent 上报数据
- 检查服务器日志

## 帮助

遇到问题？

- 查看日志输出
- 阅读文档：[README.md](README.md)
- 提交 Issue：https://github.com/jyxjjj/Monitor/issues
