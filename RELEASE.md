# 发布指南

## 自动发布（推荐）

使用 GitHub Actions 自动构建和发布：

### 1. 创建 Git 标签

```bash
# 创建并推送标签（例如 v1.0.0）
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### 2. GitHub Actions 自动执行

推送标签后，GitHub Actions 会自动：
- 构建所有架构的二进制文件（AMD64、ARM64、ARM32）
- 生成 SHA256 校验和
- 创建 GitHub Release
- 上传所有文件作为 Release Assets

### 3. 查看发布

访问 GitHub Releases 页面查看发布结果：
```
https://github.com/YOUR_USERNAME/process_exporter/releases
```

## 手动发布

如果需要手动发布：

### 1. 本地构建

```bash
# 运行构建脚本
./build.sh
```

### 2. 创建 Release

1. 访问 GitHub 仓库的 Releases 页面
2. 点击 "Draft a new release"
3. 填写版本号（例如 `v1.0.0`）
4. 填写发布说明
5. 上传 `build/` 目录下的所有文件：
   - `process_exporter_linux_amd64`
   - `process_exporter_linux_arm64`
   - `process_exporter_linux_arm32`
6. 点击 "Publish release"

### 3. 生成校验和（可选但推荐）

```bash
cd build
sha256sum process_exporter_linux_* > checksums.txt
```

然后将 `checksums.txt` 也上传到 Release。

## 版本号规范

建议使用 [语义化版本](https://semver.org/)：
- `v1.0.0` - 主版本.次版本.修订版本
- `v1.0.1` - 修复版本
- `v1.1.0` - 新功能版本
- `v2.0.0` - 重大变更版本

## 发布检查清单

- [ ] 更新版本号
- [ ] 更新 CHANGELOG.md（如果有）
- [ ] 更新 README.md（如果需要）
- [ ] 运行测试（如果有）
- [ ] 本地构建测试
- [ ] 创建 Git 标签
- [ ] 推送标签触发自动发布
- [ ] 验证 Release 文件
- [ ] 更新文档链接（如果需要）

