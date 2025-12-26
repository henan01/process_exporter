# GitHub Actions Release 故障排查

## 常见问题

### 1. "Create Release" 步骤失败

#### 问题：权限不足

**错误信息：**
```
Resource not accessible by integration
```

**解决方案：**

1. 检查工作流文件是否包含权限配置：
   ```yaml
   permissions:
     contents: write
   ```

2. 检查仓库设置：
   - 访问：`Settings > Actions > General`
   - 确保 "Workflow permissions" 设置为 "Read and write permissions"
   - 或者选择 "Read repository contents and packages permissions" 并勾选 "Allow GitHub Actions to create and approve pull requests"

#### 问题：Release 已存在

**错误信息：**
```
Release already exists
```

**解决方案：**

1. 删除已存在的 Release：
   - 访问 Releases 页面
   - 找到对应的 Release
   - 点击 "Delete release"

2. 或者使用不同的版本号

3. 或者在 workflow 中添加 `overwrite: true`（如果 action 支持）

#### 问题：找不到文件

**错误信息：**
```
No files found matching pattern
```

**解决方案：**

1. 检查 "Prepare release assets" 步骤的日志
2. 确认 artifacts 是否正确下载
3. 检查文件路径是否正确

**调试步骤：**

在 workflow 中添加调试输出：
```yaml
- name: Debug artifacts
  run: |
    echo "=== Artifacts 目录 ==="
    find artifacts -type f
    ls -la artifacts/
```

### 2. 构建步骤失败

#### 问题：Go 版本不兼容

**解决方案：**

检查 `go.mod` 中的 Go 版本要求，确保 workflow 中使用的版本匹配：
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.20'  # 确保与 go.mod 中的版本匹配
```

#### 问题：依赖下载失败

**解决方案：**

1. 检查网络连接
2. 使用 Go 代理：
   ```yaml
   - name: Set up Go
     uses: actions/setup-go@v5
     with:
       go-version: '1.20'
   - name: Configure Go proxy
     run: |
       go env -w GOPROXY=https://goproxy.cn,direct
   ```

### 3. Artifacts 下载失败

#### 问题：Artifacts 过期

**解决方案：**

在 workflow 中增加 artifacts 保留时间：
```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    retention-days: 30  # 增加保留天数
```

### 4. 文件上传失败

#### 问题：文件大小超过限制

GitHub Release 单个文件限制为 2GB，但建议小于 100MB。

**解决方案：**

1. 检查二进制文件大小
2. 使用 `-ldflags="-s -w"` 减小文件大小
3. 考虑使用压缩

## 调试技巧

### 1. 查看完整日志

在 GitHub Actions 页面：
1. 点击失败的 job
2. 展开失败的步骤
3. 查看详细错误信息

### 2. 本地测试

在本地运行类似的命令来测试：

```bash
# 模拟构建步骤
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o process_exporter_linux_amd64 main.go

# 生成校验和
sha256sum process_exporter_linux_amd64 > process_exporter_linux_amd64.sha256

# 检查文件
ls -lh process_exporter_linux_amd64*
```

### 3. 手动触发测试

使用 `workflow_dispatch` 手动触发 workflow：

1. 访问 Actions 页面
2. 选择 "Build and Release" workflow
3. 点击 "Run workflow"
4. 选择分支和版本
5. 点击 "Run workflow"

### 4. 分步调试

如果整个 workflow 失败，可以：

1. 先只运行 build job，不运行 release job
2. 手动下载 artifacts
3. 手动创建 Release 测试

## 检查清单

发布前检查：

- [ ] 工作流文件语法正确（无 YAML 错误）
- [ ] 权限配置正确（`permissions: contents: write`）
- [ ] 标签格式正确（`v1.0.0`）
- [ ] 本地构建成功
- [ ] 依赖版本匹配
- [ ] 文件大小合理
- [ ] Release 不存在（如果重新发布）

## 获取帮助

如果问题仍然存在：

1. 查看 GitHub Actions 日志的完整输出
2. 检查 [GitHub Actions 文档](https://docs.github.com/en/actions)
3. 查看 [softprops/action-gh-release 文档](https://github.com/softprops/action-gh-release)
4. 在仓库中创建 Issue 描述问题

## 常见错误代码

| 错误代码 | 含义 | 解决方案 |
|---------|------|---------|
| 403 | 权限不足 | 检查 permissions 配置 |
| 422 | 资源已存在 | 删除已存在的 Release |
| 404 | 资源不存在 | 检查标签和仓库名称 |
| 413 | 文件过大 | 减小文件大小 |

