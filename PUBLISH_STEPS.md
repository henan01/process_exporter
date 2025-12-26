# GitHub Releases 发布步骤

## 📋 准备工作

1. **确保代码已提交并推送到 GitHub**
   ```bash
   git add .
   git commit -m "准备发布 v1.0.0"
   git push origin main
   ```

2. **更新版本号（如果需要）**
   - 更新 README.md 中的版本号引用
   - 更新任何文档中的版本信息

## 🚀 发布步骤

### 方法一：使用 Git 标签自动发布（推荐）

1. **创建并推送 Git 标签**
   ```bash
   # 创建带注释的标签
   git tag -a v1.0.0 -m "Release version 1.0.0"
   
   # 推送标签到 GitHub
   git push origin v1.0.0
   ```

2. **GitHub Actions 自动执行**
   - 推送标签后，GitHub Actions 会自动触发构建
   - 构建所有架构的二进制文件（AMD64、ARM64、ARM32）
   - 生成 SHA256 校验和
   - 创建 GitHub Release
   - 上传所有文件

3. **查看发布结果**
   - 访问：`https://github.com/YOUR_USERNAME/process_exporter/releases`
   - 检查 Release 是否创建成功
   - 验证所有文件是否已上传

### 方法二：手动发布

如果自动发布失败，可以手动发布：

1. **本地构建**
   ```bash
   ./build.sh
   ```

2. **生成校验和**
   ```bash
   cd build
   sha256sum process_exporter_linux_* > checksums.txt
   ```

3. **创建 GitHub Release**
   - 访问 GitHub 仓库的 Releases 页面
   - 点击 "Draft a new release"
   - 填写版本号（例如 `v1.0.0`）
   - 填写发布说明（可以参考 `.github/workflows/release.yml` 中的模板）
   - 上传 `build/` 目录下的所有文件：
     - `process_exporter_linux_amd64`
     - `process_exporter_linux_arm64`
     - `process_exporter_linux_arm32`
     - `checksums.txt`
   - 点击 "Publish release"

## ✅ 发布后检查

- [ ] Release 已创建
- [ ] 所有架构的二进制文件已上传
- [ ] 校验和文件已上传
- [ ] 发布说明完整
- [ ] 下载链接可用
- [ ] 更新 README.md 中的下载链接（如果需要）

## 📝 发布说明模板

```markdown
## Process Exporter v1.0.0

### 📦 下载

请根据您的系统架构下载对应的二进制文件：

- **Linux AMD64 (x86_64)**: `process_exporter_linux_amd64`
- **Linux ARM64 (aarch64)**: `process_exporter_linux_arm64`
- **Linux ARM32 (ARMv7)**: `process_exporter_linux_arm32`

### 🔐 验证文件完整性

下载后可以使用 SHA256 校验和验证文件：

```bash
sha256sum -c process_exporter_linux_amd64.sha256
```

### 🚀 快速开始

```bash
# 1. 下载并添加执行权限
chmod +x process_exporter_linux_amd64

# 2. 运行
./process_exporter_linux_amd64 \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=60s \
  --top=10
```

### 📚 更多信息

详细使用说明请查看 [README.md](README.md)

### 🎉 新功能

- 功能 1
- 功能 2
- 功能 3

### 🐛 修复

- 修复 1
- 修复 2

### ⚠️ 变更

- 变更 1
- 变更 2
```

## 🔧 故障排查

### 问题：GitHub Actions 没有触发

**检查：**
- 标签格式是否正确（必须以 `v` 开头，如 `v1.0.0`）
- 标签是否已推送到 GitHub
- GitHub Actions 是否已启用（Settings > Actions > General）

### 问题：构建失败

**检查：**
- 查看 GitHub Actions 日志
- 确认 Go 版本兼容性
- 检查依赖是否正确

### 问题：文件上传失败

**检查：**
- 文件大小是否超过 GitHub 限制（100MB）
- GitHub Token 权限是否足够
- 网络连接是否正常

## 📚 相关文档

- [RELEASE.md](RELEASE.md) - 详细发布指南
- [README.md](README.md) - 项目文档
- [QUICKSTART.md](QUICKSTART.md) - 快速开始指南

