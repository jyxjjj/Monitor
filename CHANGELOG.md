# 更新日志

所有重要的项目更改都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [1.0.0] - 2025-10-16

### 新增
- 完整的服务器监控系统
- Go 语言实现的服务端和 Agent
- React + Material-UI 前端界面
- JWT 认证系统
- SQLite 数据库存储
- 基础指标收集（CPU、内存、磁盘、网络、负载）
- 历史数据存储和查询
- 实时数据可视化图表
- 告警规则系统
- 阈值触发告警
- 邮件通知功能
- TLS/HTTPS 支持
- Brotli 数据压缩
- 跨平台 Agent 支持（Linux/Windows/macOS）
- 秒级数据上报
- 心跳检测机制
- JSON 配置文件
- 自动生成默认配置
- Docker 支持
- Docker Compose 配置
- systemd 服务文件
- Makefile 构建脚本
- 跨平台编译支持
- GitHub Actions CI/CD
- 完整的文档（README、快速开始、部署指南、API 文档）
- 示例配置文件
- 安装脚本

### 技术栈
- 后端：Go 1.21+
- 数据库：SQLite
- 前端：React 18 + Material-UI 5
- 图表：MUI X-Charts
- 认证：JWT
- 压缩：Brotli
- 容器：Docker

### API 端点
- POST /api/login - 管理员登录
- GET /api/agents - 获取 Agent 列表
- GET /api/metrics/{agentId} - 获取指标历史
- POST /api/metrics/report - Agent 上报数据
- GET /api/alerts - 获取告警列表
- GET /api/alert-rules - 获取告警规则
- POST /api/alert-rules - 创建告警规则
- GET /api/config - 获取配置信息

## [未发布] - 开发中

### 计划添加
- WebSocket 实时推送
- 多用户支持
- 更多指标类型
- 数据导出功能
- Grafana 集成

---

[1.0.0]: https://github.com/jyxjjj/Monitor/releases/tag/v1.0.0
