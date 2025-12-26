#!/bin/bash

# 构建脚本 - VmAgent Process Exporter
# 用于编译 Linux 多架构版本（amd64, arm64, arm32）

set -e

echo "开始编译 VmAgent Process Exporter..."
echo ""

# 设置公共编译参数
export CGO_ENABLED=0
LDFLAGS="-s -w"

# 创建 build 目录
mkdir -p build

# 编译 AMD64 版本
echo "==> 编译 Linux AMD64 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/process_exporter_linux_amd64 main.go
echo "✓ AMD64 编译完成: build/process_exporter_linux_amd64"
echo ""

# 编译 ARM64 版本
echo "==> 编译 Linux ARM64 版本..."
GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o build/process_exporter_linux_arm64 main.go
echo "✓ ARM64 编译完成: build/process_exporter_linux_arm64"
echo ""

# 编译 ARM32 版本（ARMv7）
echo "==> 编译 Linux ARM32 版本..."
GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$LDFLAGS" -o build/process_exporter_linux_arm32 main.go
echo "✓ ARM32 编译完成: build/process_exporter_linux_arm32"
echo ""

echo "=========================================="
echo "所有版本编译完成！"
echo "=========================================="
echo ""
echo "编译产物："
ls -lh build/
echo ""
echo "使用方法："
echo "  AMD64: ./build/process_exporter_linux_amd64 --remote.url=http://victoriametrics:8428/api/v1/write"
echo "  ARM64: ./build/process_exporter_linux_arm64 --remote.url=http://victoriametrics:8428/api/v1/write"
echo "  ARM32: ./build/process_exporter_linux_arm32 --remote.url=http://victoriametrics:8428/api/v1/write"
echo ""
echo "查看所有参数:"
echo "  ./build/process_exporter_linux_amd64 --help"
