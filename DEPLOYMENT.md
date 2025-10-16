# Monitor 部署指南

本文档详细说明如何部署 Monitor 监控系统。

## 目录

- [系统要求](#系统要求)
- [安装方式](#安装方式)
  - [方式一：直接安装](#方式一直接安装)
  - [方式二：Docker 部署](#方式二docker-部署)
  - [方式三：系统服务部署](#方式三系统服务部署)
- [配置说明](#配置说明)
- [TLS 配置](#tls-配置)
- [常见问题](#常见问题)

## 系统要求

### 服务器端
- Linux/Windows/macOS
- 512MB RAM (最小)
- 1GB 磁盘空间
- Go 1.21+ (编译需要)
- Node.js 16+ (构建前端需要)

### Agent 端
- Linux/Windows/macOS
- 最小资源占用：10MB RAM, 20MB 磁盘

## 安装方式

### 方式一：直接安装

#### 1. 克隆仓库

```bash
git clone https://github.com/jyxjjj/Monitor.git
cd Monitor
```

#### 2. 构建

```bash
# 安装依赖
go mod download

# 构建服务器和 Agent
make build

# (可选) 构建前端
cd frontend
npm install
npm run build
cd ..
```

#### 3. 配置

首次运行会自动生成配置文件：

```bash
# 运行服务器生成配置
./monitor-server

# 运行 Agent 生成配置
./monitor-agent
```

编辑生成的配置文件：
- `server-config.json` - 服务器配置
- `agent-config.json` - Agent 配置

#### 4. 运行

```bash
# 启动服务器
./monitor-server

# 在其他服务器上启动 Agent
./monitor-agent
```

### 方式二：Docker 部署

#### 1. 准备配置文件

```bash
# 复制示例配置
cp server-config.example.json server-config.json
cp agent-config.example.json agent-config.json

# 编辑配置文件
vim server-config.json
vim agent-config.json
```

#### 2. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 3. 仅部署 Agent

如果只需要部署 Agent 到其他服务器：

```bash
docker run -d \
  --name monitor-agent \
  -v /path/to/agent-config.json:/app/agent-config.json \
  --restart unless-stopped \
  monitor-agent
```

### 方式三：系统服务部署

#### 在 Linux 上使用 systemd

1. **创建用户和目录**

```bash
# 创建系统用户
sudo useradd -r -s /bin/false monitor

# 创建目录
sudo mkdir -p /opt/monitor
sudo mkdir -p /etc/monitor
sudo mkdir -p /var/lib/monitor

# 设置权限
sudo chown -R monitor:monitor /opt/monitor
sudo chown -R monitor:monitor /var/lib/monitor
```

2. **安装文件**

```bash
# 复制二进制文件
sudo cp monitor-server /opt/monitor/
sudo cp monitor-agent /opt/monitor/

# 复制配置文件
sudo cp server-config.json /etc/monitor/
sudo cp agent-config.json /etc/monitor/

# 设置权限
sudo chmod 755 /opt/monitor/monitor-*
sudo chmod 644 /etc/monitor/*.json
```

3. **安装 systemd 服务**

```bash
# 复制服务文件
sudo cp scripts/monitor-server.service /etc/systemd/system/
sudo cp scripts/monitor-agent.service /etc/systemd/system/

# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start monitor-server
sudo systemctl start monitor-agent

# 设置开机自启
sudo systemctl enable monitor-server
sudo systemctl enable monitor-agent

# 查看状态
sudo systemctl status monitor-server
sudo systemctl status monitor-agent
```

4. **查看日志**

```bash
# 服务器日志
sudo journalctl -u monitor-server -f

# Agent 日志
sudo journalctl -u monitor-agent -f
```

## 配置说明

### 服务器配置 (server-config.json)

```json
{
  "server_addr": ":8443",           // 监听地址和端口
  "tls_cert_file": "server.crt",    // TLS 证书文件路径 (可选)
  "tls_key_file": "server.key",     // TLS 密钥文件路径 (可选)
  "db_path": "./monitor.db",        // 数据库文件路径
  "admin_password": "admin123",     // 管理员密码 (请修改)
  "smtp_host": "smtp.gmail.com",    // SMTP 服务器
  "smtp_port": 587,                 // SMTP 端口
  "smtp_user": "user@gmail.com",    // SMTP 用户名
  "smtp_password": "password",      // SMTP 密码
  "email_from": "user@gmail.com",   // 发件人邮箱
  "alert_email": "alert@example.com" // 告警接收邮箱
}
```

### Agent 配置 (agent-config.json)

```json
{
  "server_url": "https://monitor.example.com:8443", // 服务器地址
  "agent_id": "server-01",          // Agent ID (唯一)
  "agent_name": "Production Server 01", // Agent 名称
  "report_interval": 5,             // 上报间隔 (秒)
  "tls_skip_verify": false          // 是否跳过 TLS 验证
}
```

### 重要配置项说明

1. **admin_password**: 强烈建议修改为强密码
2. **server_addr**: 如果使用反向代理，可以使用 `:8080`
3. **tls_cert_file/tls_key_file**: 生产环境建议启用 HTTPS
4. **report_interval**: 建议值 5-60 秒
5. **tls_skip_verify**: 生产环境应设置为 `false`

## TLS 配置

### 生成自签名证书（开发/测试）

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes -subj "/CN=localhost"
```

### 使用 Let's Encrypt（生产环境）

```bash
# 安装 certbot
sudo apt install certbot

# 获取证书
sudo certbot certonly --standalone -d monitor.example.com

# 证书位置
# /etc/letsencrypt/live/monitor.example.com/fullchain.pem
# /etc/letsencrypt/live/monitor.example.com/privkey.pem
```

更新配置：

```json
{
  "tls_cert_file": "/etc/letsencrypt/live/monitor.example.com/fullchain.pem",
  "tls_key_file": "/etc/letsencrypt/live/monitor.example.com/privkey.pem"
}
```

### 使用 Nginx 反向代理

如果不想在应用层配置 TLS，可以使用 Nginx：

```nginx
server {
    listen 443 ssl http2;
    server_name monitor.example.com;

    ssl_certificate /etc/letsencrypt/live/monitor.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/monitor.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

服务器配置：

```json
{
  "server_addr": ":8080",
  "tls_cert_file": "",
  "tls_key_file": ""
}
```

## 邮件告警配置

### Gmail 配置

1. 启用两步验证
2. 生成应用专用密码：https://myaccount.google.com/apppasswords
3. 配置：

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

### 其他邮件服务

常见 SMTP 配置：

- **Outlook**: smtp-mail.outlook.com:587
- **QQ邮箱**: smtp.qq.com:587
- **163邮箱**: smtp.163.com:465

## 常见问题

### 1. 服务器无法启动

**问题**: `bind: address already in use`

**解决**: 端口被占用，修改 `server_addr` 或停止占用端口的进程

```bash
# 查看端口占用
sudo lsof -i :8443
sudo netstat -tulpn | grep 8443
```

### 2. Agent 无法连接服务器

**问题**: `connection refused` 或 `timeout`

**解决**:
- 检查服务器是否正常运行
- 检查防火墙设置
- 验证 `server_url` 配置正确

```bash
# 测试连接
curl -v https://your-server:8443/api/agents
```

### 3. 数据库错误

**问题**: `database is locked`

**解决**: 
- 确保只有一个服务器实例运行
- 检查数据库文件权限

### 4. 前端无法访问

**问题**: 页面显示空白或 404

**解决**:
- 确认前端已构建：`cd frontend && npm run build`
- 检查浏览器控制台错误
- 验证 API 路径配置

### 5. TLS 证书错误

**问题**: `x509: certificate signed by unknown authority`

**解决**:
- 使用有效的证书
- 开发环境可临时设置 `tls_skip_verify: true`
- 确保证书包含正确的域名

## 性能优化

### 数据清理

定期清理旧数据以保持性能：

```bash
# 添加到 crontab
0 2 * * * sqlite3 /var/lib/monitor/monitor.db "DELETE FROM metrics WHERE timestamp < datetime('now', '-30 days');"
```

### 数据库优化

```bash
# 定期执行 VACUUM
sqlite3 /var/lib/monitor/monitor.db "VACUUM;"
```

### 调整上报间隔

根据需要调整 Agent 的 `report_interval`：
- 实时监控: 5 秒
- 一般监控: 30 秒
- 低频监控: 60 秒

## 安全建议

1. ✅ 使用强密码
2. ✅ 启用 HTTPS/TLS
3. ✅ 定期更新软件
4. ✅ 限制服务器访问（防火墙/IP 白名单）
5. ✅ 定期备份数据库
6. ✅ 监控系统日志

## 备份和恢复

### 备份

```bash
# 备份数据库
cp /var/lib/monitor/monitor.db /backup/monitor-$(date +%Y%m%d).db

# 备份配置
tar czf /backup/monitor-config-$(date +%Y%m%d).tar.gz /etc/monitor/
```

### 恢复

```bash
# 停止服务
sudo systemctl stop monitor-server

# 恢复数据库
cp /backup/monitor-20231215.db /var/lib/monitor/monitor.db

# 启动服务
sudo systemctl start monitor-server
```

## 支持

如有问题，请访问：
- GitHub Issues: https://github.com/jyxjjj/Monitor/issues
- 文档: https://github.com/jyxjjj/Monitor
