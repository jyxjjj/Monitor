# Monitor API 文档

本文档描述 Monitor 的 REST API 接口。

## 基本信息

- **Base URL**: `http://your-server:8443`
- **Content-Type**: `application/json`
- **认证方式**: JWT Bearer Token

## 认证

### 登录

获取访问令牌。

**请求**

```
POST /api/login
```

**请求体**

```json
{
  "password": "admin123"
}
```

**响应**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**示例**

```bash
curl -X POST http://localhost:8443/api/login \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'
```

## Agent 管理

所有 Agent 相关的接口都需要认证。

### 获取 Agent 列表

**请求**

```
GET /api/agents
```

**Headers**

```
Authorization: Bearer YOUR_TOKEN
```

**响应**

```json
[
  {
    "id": "server-01",
    "name": "Production Server 01",
    "host": "192.168.1.100:45678",
    "last_seen": "2023-12-15T10:30:00Z",
    "status": "online",
    "platform": "linux",
    "version": "1.0.0"
  }
]
```

**状态说明**

- `online`: Agent 在线（最近 2 分钟有心跳）
- `offline`: Agent 离线

**示例**

```bash
TOKEN="your_jwt_token"
curl http://localhost:8443/api/agents \
  -H "Authorization: Bearer $TOKEN"
```

## 指标数据

### 获取指标历史

获取指定 Agent 的历史监控数据。

**请求**

```
GET /api/metrics/{agentId}?since={timestamp}
```

**参数**

- `agentId`: Agent ID
- `since`: 起始时间（ISO 8601 格式，可选）

**Headers**

```
Authorization: Bearer YOUR_TOKEN
```

**响应**

```json
[
  {
    "agent_id": "server-01",
    "timestamp": "2023-12-15T10:30:00Z",
    "cpu_percent": 25.5,
    "memory_used": 8589934592,
    "memory_total": 17179869184,
    "disk_used": 107374182400,
    "disk_total": 536870912000,
    "network_rx": 1024000,
    "network_tx": 512000,
    "load_avg_1": 1.5,
    "load_avg_5": 1.2,
    "load_avg_15": 1.0
  }
]
```

**示例**

```bash
# 获取最近 1 小时的数据
SINCE=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
curl "http://localhost:8443/api/metrics/server-01?since=$SINCE" \
  -H "Authorization: Bearer $TOKEN"
```

### Agent 上报数据

Agent 使用此接口上报监控数据。

**请求**

```
POST /api/metrics/report
```

**Headers**

```
Content-Type: application/json
Content-Encoding: br (可选，Brotli 压缩)
```

**请求体**

```json
{
  "agent_id": "server-01",
  "timestamp": "2023-12-15T10:30:00Z",
  "cpu_percent": 25.5,
  "memory_used": 8589934592,
  "memory_total": 17179869184,
  "disk_used": 107374182400,
  "disk_total": 536870912000,
  "network_rx": 1024000,
  "network_tx": 512000,
  "load_avg_1": 1.5,
  "load_avg_5": 1.2,
  "load_avg_15": 1.0
}
```

**响应**

```
200 OK
```

## 告警规则

### 获取告警规则列表

**请求**

```
GET /api/alert-rules
```

**Headers**

```
Authorization: Bearer YOUR_TOKEN
```

**响应**

```json
[
  {
    "id": 1,
    "agent_id": "server-01",
    "metric_type": "cpu",
    "threshold": 80,
    "operator": "gt",
    "duration": 60,
    "enabled": true,
    "description": "CPU usage exceeds 80%"
  }
]
```

**字段说明**

- `metric_type`: cpu | memory | disk | load
- `operator`: gt (>) | lt (<) | gte (>=) | lte (<=)
- `threshold`: 阈值（百分比或数值）
- `duration`: 持续时间（秒）

**示例**

```bash
curl http://localhost:8443/api/alert-rules \
  -H "Authorization: Bearer $TOKEN"
```

### 创建告警规则

**请求**

```
POST /api/alert-rules
```

**Headers**

```
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json
```

**请求体**

```json
{
  "agent_id": "server-01",
  "metric_type": "cpu",
  "threshold": 80,
  "operator": "gt",
  "duration": 60,
  "enabled": true,
  "description": "CPU usage exceeds 80%"
}
```

**响应**

```json
{
  "id": 1,
  "agent_id": "server-01",
  "metric_type": "cpu",
  "threshold": 80,
  "operator": "gt",
  "duration": 60,
  "enabled": true,
  "description": "CPU usage exceeds 80%"
}
```

**示例**

```bash
curl -X POST http://localhost:8443/api/alert-rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "server-01",
    "metric_type": "memory",
    "threshold": 90,
    "operator": "gt",
    "duration": 120,
    "enabled": true,
    "description": "Memory usage exceeds 90%"
  }'
```

## 告警历史

### 获取告警列表

获取最近的告警记录。

**请求**

```
GET /api/alerts
```

**Headers**

```
Authorization: Bearer YOUR_TOKEN
```

**响应**

```json
[
  {
    "id": 1,
    "rule_id": 1,
    "agent_id": "server-01",
    "timestamp": "2023-12-15T10:30:00Z",
    "message": "cpu: 85.50% gt 80.00%",
    "value": 85.5,
    "resolved": false
  }
]
```

**示例**

```bash
curl http://localhost:8443/api/alerts \
  -H "Authorization: Bearer $TOKEN"
```

## 配置

### 获取服务器配置

获取非敏感的服务器配置信息。

**请求**

```
GET /api/config
```

**Headers**

```
Authorization: Bearer YOUR_TOKEN
```

**响应**

```json
{
  "server_addr": ":8443",
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "email_from": "monitor@example.com",
  "alert_email": "admin@example.com"
}
```

**示例**

```bash
curl http://localhost:8443/api/config \
  -H "Authorization: Bearer $TOKEN"
```

## 错误响应

所有错误响应遵循以下格式：

**响应**

```
HTTP/1.1 4xx/5xx
Content-Type: text/plain

Error message
```

**常见错误代码**

- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权或 token 无效
- `404 Not Found`: 资源不存在
- `405 Method Not Allowed`: HTTP 方法不允许
- `500 Internal Server Error`: 服务器内部错误

## 数据压缩

Agent 上报数据支持 Brotli 压缩，在请求头中添加：

```
Content-Encoding: br
```

服务器会自动解压缩数据。

## 认证令牌

JWT token 有效期为 24 小时。过期后需要重新登录获取新的 token。

## 速率限制

当前版本没有实现速率限制，建议在反向代理层面实现。

## WebSocket (未实现)

未来版本可能会添加 WebSocket 支持以实现实时数据推送。

## 示例脚本

### Python 客户端示例

```python
import requests
import json

class MonitorClient:
    def __init__(self, base_url, password):
        self.base_url = base_url
        self.token = self.login(password)
    
    def login(self, password):
        response = requests.post(
            f"{self.base_url}/api/login",
            json={"password": password}
        )
        return response.json()["token"]
    
    def get_agents(self):
        response = requests.get(
            f"{self.base_url}/api/agents",
            headers={"Authorization": f"Bearer {self.token}"}
        )
        return response.json()
    
    def get_metrics(self, agent_id, since=None):
        url = f"{self.base_url}/api/metrics/{agent_id}"
        if since:
            url += f"?since={since}"
        response = requests.get(
            url,
            headers={"Authorization": f"Bearer {self.token}"}
        )
        return response.json()

# 使用示例
client = MonitorClient("http://localhost:8443", "admin123")
agents = client.get_agents()
print(f"Found {len(agents)} agents")
```

### Shell 脚本示例

```bash
#!/bin/bash

BASE_URL="http://localhost:8443"
PASSWORD="admin123"

# 登录
TOKEN=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "{\"password\":\"$PASSWORD\"}" | jq -r .token)

# 获取 Agent 列表
curl -s "$BASE_URL/api/agents" \
  -H "Authorization: Bearer $TOKEN" | jq .

# 获取指标
curl -s "$BASE_URL/api/metrics/server-01" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## 变更日志

### v1.0.0
- 初始版本
- 基本的 API 端点
- JWT 认证
- Brotli 压缩支持
