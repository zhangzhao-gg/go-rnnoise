# 贡献指南

感谢您对 Go RNNoise 项目的关注！我们欢迎各种形式的贡献，包括但不限于：

- 🐛 Bug 报告
- 💡 功能建议
- 🔧 代码贡献
- 📖 文档改进
- 🧪 测试用例

## 开发环境设置

### 前置要求

- Go 1.19 或更高版本
- C 编译器（用于 CGO）
- Git

### 克隆项目

```bash
git clone https://github.com/zhangzhao-gg/go-rnnoise.git
cd go-rnnoise
```

### 安装依赖

```bash
make deps
```

### 安装开发工具

```bash
make install-tools
```

## 开发流程

### 1. Fork 和分支

1. Fork 这个仓库
2. 创建您的特性分支：`git checkout -b feature/amazing-feature`
3. 提交您的更改：`git commit -m 'Add some amazing feature'`
4. 推送到分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

### 2. 代码规范

#### Go 代码规范

- 遵循 [Go 官方代码规范](https://golang.org/doc/effective_go.html)
- 使用 `gofmt` 格式化代码
- 使用 `goimports` 整理导入
- 为所有公共 API 添加文档注释

#### 运行代码检查

```bash
make check  # 运行格式化、linting、测试和安全检查
```

或者分别运行：

```bash
make fmt     # 格式化代码
make lint    # 运行 linter
make test    # 运行测试
make security # 安全检查
```

### 3. 测试

#### 运行测试

```bash
make test
```

#### 运行测试并生成覆盖率报告

```bash
make test-coverage
```

#### 运行示例

```bash
make demo
```

### 4. 文档

- 为所有公共 API 添加 Go 文档注释
- 更新 README.md 如果添加了新功能
- 更新 CHANGELOG.md 记录重要更改

## 提交规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### 类型 (type)

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更改
- `style`: 代码格式更改（不影响功能）
- `refactor`: 代码重构
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

### 示例

```
feat(audio): add support for 24-bit audio processing

- Add ConvertBytesToFloat32 support for 24-bit depth
- Update documentation with new format examples

Closes #123
```

## Issue 报告

在提交 Issue 之前，请确保：

1. 检查是否已存在相同的 Issue
2. 使用清晰、描述性的标题
3. 提供详细的复现步骤
4. 包含环境信息（操作系统、Go 版本等）
5. 如果是 Bug，请提供错误日志

### Bug 报告模板

```markdown
## Bug 描述
简要描述发现的 Bug。

## 复现步骤
1. 执行 '...'
2. 点击 '....'
3. 滚动到 '....'
4. 看到错误

## 期望行为
描述您期望发生的情况。

## 实际行为
描述实际发生的情况。

## 环境信息
- 操作系统: [e.g. macOS 12.0]
- Go 版本: [e.g. 1.21.0]
- 项目版本: [e.g. v1.0.0]

## 附加信息
添加任何其他相关的上下文信息。
```

## Pull Request 指南

### PR 标题

使用清晰的标题，描述 PR 的主要内容：

```
feat: add support for batch audio processing
fix: resolve memory leak in audio conversion
docs: update API documentation
```

### PR 描述

包含以下信息：

1. **变更摘要**: 简要描述这个 PR 做了什么
2. **变更类型**: Bug 修复、新功能、文档更新等
3. **测试**: 描述如何测试这些更改
4. **检查列表**: 确保 PR 质量

### PR 检查列表

- [ ] 代码遵循项目规范
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 所有测试通过
- [ ] 代码通过了 linting 检查
- [ ] 更新了 CHANGELOG.md（如适用）

## 发布流程

### 版本号规范

我们遵循 [语义化版本](https://semver.org/) 规范：

- `MAJOR`: 不兼容的 API 更改
- `MINOR`: 向后兼容的功能添加
- `PATCH`: 向后兼容的 Bug 修复

### 发布检查列表

- [ ] 更新版本号
- [ ] 更新 CHANGELOG.md
- [ ] 更新 README.md（如需要）
- [ ] 运行完整测试套件
- [ ] 构建多平台二进制文件
- [ ] 创建 GitHub Release

## 社区行为准则

请遵守我们的 [行为准则](CODE_OF_CONDUCT.md)，创造一个友好和包容的环境。

## 联系方式

- 项目维护者: [@zhangzhao-gg](https://github.com/zhangzhao-gg)
- 项目主页: [https://github.com/zhangzhao-gg/go-rnnoise](https://github.com/zhangzhao-gg/go-rnnoise)

感谢您的贡献！🎉
