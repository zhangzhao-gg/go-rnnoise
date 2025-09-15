# 发布指南

本文档描述了如何为 Go RNNoise 项目创建和发布新版本。

## 发布流程

### 1. 准备发布

#### 检查清单

- [ ] 所有功能已完成并经过测试
- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 版本号已确定

#### 运行完整测试

```bash
make check
```

这包括：
- 代码格式化检查
- Linting 检查
- 单元测试
- 安全漏洞检查

#### 构建多平台二进制文件

```bash
make build-all
```

### 2. 更新版本信息

#### 更新 go.mod

确保 go.mod 中的模块版本正确。

#### 更新 CHANGELOG.md

在 CHANGELOG.md 中添加新版本的条目：

```markdown
## [1.1.0] - 2024-01-XX

### Added
- 新功能描述

### Changed
- 变更描述

### Fixed
- Bug 修复描述
```

### 3. 创建 Git 标签

```bash
# 创建带注释的标签
git tag -a v1.1.0 -m "Release version 1.1.0"

# 推送标签到远程仓库
git push origin v1.1.0
```

### 4. 创建 GitHub Release

1. 访问 [GitHub Releases 页面](https://github.com/zhangzhao-gg/go-rnnoise/releases)
2. 点击 "Create a new release"
3. 选择刚创建的标签
4. 填写发布标题和描述
5. 上传构建的二进制文件
6. 发布

### 5. 验证发布

#### 测试安装

```bash
# 测试从 GitHub 安装
go get github.com/zhangzhao-gg/go-rnnoise@v1.1.0

# 验证版本
go list -m github.com/zhangzhao-gg/go-rnnoise
```

#### 测试示例

```bash
# 构建并测试示例
cd example
go build -o rnnoise-cli .
./rnnoise-cli --help
```

## 版本号规范

我们遵循 [语义化版本](https://semver.org/) 规范：

- **MAJOR** (主版本号): 当进行不兼容的 API 更改
- **MINOR** (次版本号): 当以向后兼容的方式添加功能
- **PATCH** (修订号): 当进行向后兼容的 Bug 修复

### 版本号示例

- `1.0.0` - 初始稳定版本
- `1.0.1` - Bug 修复
- `1.1.0` - 新功能添加
- `2.0.0` - 重大更改（破坏向后兼容性）

## 发布类型

### 主要版本 (Major Release)

- 包含破坏性更改
- 需要更新用户代码
- 详细记录迁移指南

### 次要版本 (Minor Release)

- 添加新功能
- 向后兼容
- 更新文档和示例

### 补丁版本 (Patch Release)

- Bug 修复
- 性能改进
- 文档修正

## 发布后任务

### 1. 更新文档

- [ ] 更新 README.md 中的版本信息
- [ ] 更新示例代码
- [ ] 更新安装说明

### 2. 社区通知

- [ ] 在相关论坛发布公告
- [ ] 更新项目状态
- [ ] 回复用户问题

### 3. 监控

- [ ] 监控下载统计
- [ ] 收集用户反馈
- [ ] 跟踪问题报告

## 回滚计划

如果发布后发现问题：

### 1. 立即响应

- 评估问题严重程度
- 决定是否需要回滚
- 通知用户

### 2. 创建热修复

- 快速修复问题
- 创建新的补丁版本
- 重新发布

### 3. 标记问题版本

- 在 GitHub 上标记问题
- 更新发布说明
- 提供迁移指导

## 自动化发布

考虑使用 GitHub Actions 自动化发布流程：

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Create Release
        uses: actions/create-release@v1
        # ... 配置发布步骤
```

## 发布检查清单

### 发布前

- [ ] 代码审查完成
- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG 已更新
- [ ] 版本号已确定
- [ ] 标签已创建

### 发布中

- [ ] GitHub Release 已创建
- [ ] 二进制文件已上传
- [ ] 发布说明完整

### 发布后

- [ ] 安装测试通过
- [ ] 示例运行正常
- [ ] 社区通知已发送
- [ ] 监控已设置

## 故障排除

### 常见问题

1. **标签推送失败**
   ```bash
   git push origin v1.1.0
   ```

2. **构建失败**
   - 检查 Go 版本
   - 验证依赖项
   - 清理构建缓存

3. **发布失败**
   - 检查 GitHub 权限
   - 验证标签存在
   - 重试发布流程

### 获取帮助

如果遇到问题，请：

1. 查看 GitHub Issues
2. 联系维护者
3. 查看项目文档
