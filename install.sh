#!/bin/bash

# Process Exporter 快速安装脚本
# 从 GitHub Releases 下载并安装

set -e

# 配置
REPO="YOUR_USERNAME/process_exporter"  # 请替换为您的 GitHub 仓库
VERSION="${1:-latest}"  # 版本号，默认为 latest
INSTALL_DIR="${2:-/usr/local/bin}"  # 安装目录

# 检测系统架构
detect_arch() {
    case $(uname -m) in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        armv7l|armv6l)
            ARCH="arm32"
            ;;
        *)
            echo "错误: 不支持的架构 $(uname -m)"
            exit 1
            ;;
    esac
}

# 获取最新版本号
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    echo "使用版本: $VERSION"
}

# 下载文件
download_file() {
    local file=$1
    local url="https://github.com/${REPO}/releases/download/${VERSION}/${file}"
    
    echo "下载: $url"
    if ! curl -L -f -o "$file" "$url"; then
        echo "错误: 下载失败 $file"
        exit 1
    fi
}

# 主函数
main() {
    echo "=========================================="
    echo "Process Exporter 安装脚本"
    echo "=========================================="
    echo ""
    
    # 检测架构
    detect_arch
    echo "检测到架构: $ARCH"
    
    # 获取版本
    get_latest_version
    
    # 设置文件名
    BINARY="process_exporter_linux_${ARCH}"
    CHECKSUM="${BINARY}.sha256"
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # 下载二进制文件
    echo ""
    echo "下载二进制文件..."
    download_file "$BINARY"
    
    # 下载校验和文件
    echo "下载校验和文件..."
    if download_file "$CHECKSUM" 2>/dev/null; then
        echo "验证文件完整性..."
        sha256sum -c "$CHECKSUM" || {
            echo "警告: 文件校验失败，但继续安装..."
        }
    else
        echo "警告: 无法下载校验和文件，跳过验证"
    fi
    
    # 安装
    echo ""
    echo "安装到 $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR"
    sudo cp "$BINARY" "$INSTALL_DIR/process_exporter"
    sudo chmod +x "$INSTALL_DIR/process_exporter"
    
    # 清理
    cd -
    rm -rf "$TEMP_DIR"
    
    # 验证安装
    echo ""
    echo "验证安装..."
    if command -v process_exporter >/dev/null 2>&1; then
        echo "✓ 安装成功！"
        echo ""
        echo "使用方法:"
        echo "  process_exporter --remote.url=http://victoriametrics:8428/api/v1/write"
        echo ""
        echo "查看帮助:"
        echo "  process_exporter --help"
    else
        echo "警告: 安装可能失败，请检查 $INSTALL_DIR 是否在 PATH 中"
    fi
}

# 运行主函数
main

