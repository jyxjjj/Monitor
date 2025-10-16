# 贡献指南

感谢你考虑为 Monitor 项目做出贡献！

## 如何贡献

### 报告 Bug

如果你发现了 bug，请创建一个 issue，包含以下信息：

- Bug 的详细描述
- 复现步骤
- 期望的行为
- 实际的行为
- 环境信息（操作系统、Go 版本等）
- 相关日志

### 提出新功能

如果你有新功能的想法，请创建一个 issue，描述：

- 功能的用途
- 为什么需要这个功能
- 如何实现（可选）

### 提交代码

1. Fork 这个仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建一个 Pull Request

## 开发环境设置

### 要求

- Go 1.21+
- Node.js 16+ (用于前端开发)
- Git
- Make (可选)

### 设置步骤

```bash
# 克隆你的 fork
git clone https://github.com/YOUR_USERNAME/Monitor.git
cd Monitor

# 添加上游仓库
git remote add upstream https://github.com/jyxjjj/Monitor.git

# 安装依赖
go mod download
cd frontend && npm install && cd ..

# 构建
make build
```

## 代码规范

### Go 代码

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 添加适当的注释
- 编写测试

```bash
# 格式化代码
go fmt ./...

# 检查代码
go vet ./...

# 运行测试
go test -v ./...
```

### JavaScript/React 代码

- 使用 2 空格缩进
- 遵循 ESLint 规则
- 使用函数式组件和 Hooks
- 添加适当的 PropTypes

```bash
# 格式化和检查
cd frontend
npm run lint
```

## 提交信息规范

使用清晰的提交信息：

```
类型: 简短描述

详细描述（可选）

相关 issue: #123
```

类型：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具相关

示例：
```
feat: 添加磁盘 I/O 监控

- 添加磁盘读写速率收集
- 更新前端图表显示
- 添加相关测试

相关 issue: #45
```

## Pull Request 指南

好的 PR 应该：

1. 只包含相关的更改
2. 包含测试
3. 更新相关文档
4. 通过所有 CI 检查
5. 有清晰的描述

PR 描述模板：

```markdown
## 更改内容

简要描述这个 PR 的更改内容

## 相关 Issue

Fixes #123

## 更改类型

- [ ] Bug 修复
- [ ] 新功能
- [ ] 重构
- [ ] 文档更新
- [ ] 其他

## 检查清单

- [ ] 代码已格式化
- [ ] 添加了测试
- [ ] 更新了文档
- [ ] 通过了所有测试
```

## 测试

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定包的测试
go test -v ./pkg/collector

# 运行覆盖率测试
go test -cover ./...
```

### 编写测试

为新功能添加测试：

```go
func TestCollectorCPU(t *testing.T) {
    collector := NewCollector("test-agent")
    metrics, err := collector.Collect()
    
    if err != nil {
        t.Fatalf("Failed to collect metrics: %v", err)
    }
    
    if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
        t.Errorf("Invalid CPU percent: %f", metrics.CPUPercent)
    }
}
```

## 文档

更新文档当：

- 添加新功能
- 更改 API
- 更改配置选项
- 修复重要 bug

需要更新的文档：
- README.md
- API.md
- DEPLOYMENT.md
- 代码注释

## 发布流程

（仅维护者）

1. 更新版本号
2. 更新 CHANGELOG.md
3. 创建 git tag
4. 推送 tag
5. GitHub Actions 自动构建和发布

## 行为准则

- 尊重所有贡献者
- 接受建设性批评
- 专注于对项目最有利的事情
- 帮助新贡献者

## 问题？

如有任何问题，请：

- 创建一个 issue
- 在现有 issue 中评论
- 联系维护者

## 许可证

通过贡献，你同意你的贡献将在 MIT 许可证下授权。

感谢你的贡献！ 🎉
