# 快速开始指南

## 方式一：使用预编译版本（推荐）

### 1. 下载

访问 [GitHub Releases](https://github.com/YOUR_USERNAME/process_exporter/releases) 下载对应架构的二进制文件。

**使用 wget 下载：**

```bash
# AMD64 (x86_64)
wget https://github.com/YOUR_USERNAME/process_exporter/releases/download/v1.0.0/process_exporter_linux_amd64

# ARM64 (aarch64)
wget https://github.com/YOUR_USERNAME/process_exporter/releases/download/v1.0.0/process_exporter_linux_arm64

# ARM32 (ARMv7)
wget https://github.com/YOUR_USERNAME/process_exporter/releases/download/v1.0.0/process_exporter_linux_arm32
```

**使用 curl 下载：**

```bash
# AMD64
curl -L -o process_exporter_linux_amd64 \
  https://github.com/YOUR_USERNAME/process_exporter/releases/download/v1.0.0/process_exporter_linux_amd64
```

### 2. 验证文件完整性（可选但推荐）

```bash
# 下载校验和文件
wget https://github.com/YOUR_USERNAME/process_exporter/releases/download/v1.0.0/process_exporter_linux_amd64.sha256

# 验证
sha256sum -c process_exporter_linux_amd64.sha256
```

### 3. 添加执行权限

```bash
chmod +x process_exporter_linux_amd64
```

### 4. 运行

```bash
./process_exporter_linux_amd64 \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=60s \
  --top=10
```

## 方式二：使用安装脚本

```bash
# 自动下载并安装到 /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/process_exporter/main/install.sh | bash

# 安装后直接使用
process_exporter --remote.url=http://victoriametrics:8428/api/v1/write
```

## 方式三：从源码编译

### 1. 克隆仓库

```bash
git clone https://github.com/YOUR_USERNAME/process_exporter.git
cd process_exporter
```

### 2. 编译

```bash
# 使用构建脚本（推荐）
chmod +x build.sh
./build.sh

# 或手动编译
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o process_exporter_linux_amd64 main.go
```

### 3. 运行

```bash
./build/process_exporter_linux_amd64 \
  --remote.url=http://victoriametrics:8428/api/v1/write
```

## 配置示例

### 推送到 VictoriaMetrics

```bash
./process_exporter_linux_amd64 \
  --remote.url=http://victoriametrics:8428/api/v1/write \
  --interval=60s \
  --top=10
```

### 推送到 Prometheus

```bash
./process_exporter_linux_amd64 \
  --remote.url=http://prometheus:9090/api/v1/write \
  --interval=60s \
  --top=10
```

### 带认证的配置

```bash
./process_exporter_linux_amd64 \
  --remote.url=https://vm.example.com/api/v1/write \
  --remote.username=monitor \
  --remote.password=secret \
  --interval=60s \
  --top=10 \
  --label env=production \
  --label region=us-east-1
```

## 下一步

- 查看 [README.md](README.md) 了解详细配置选项
- 查看 [部署指南](README.md#部署) 了解如何配置为系统服务
- 查看 [查询示例](README.md#查询示例) 了解如何查询指标

