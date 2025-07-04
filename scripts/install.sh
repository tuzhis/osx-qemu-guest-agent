#!/bin/bash

# macOS Guest Agent 自动安装脚本
# Auto Installation Script for macOS Guest Agent

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 仓库信息
REPO="tuzhis/osx-qemu-guest-agent"
BINARY_NAME="mac-guest-agent"
INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "此脚本需要 root 权限运行。请使用 sudo 执行："
        print_error "sudo $0"
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            print_error "不支持的架构: $arch"
            print_error "支持的架构: x86_64 (Intel), arm64 (Apple Silicon)"
            exit 1
            ;;
    esac
}

# 检测 macOS 版本
check_macos_version() {
    local version=$(sw_vers -productVersion)
    local major=$(echo $version | cut -d. -f1)
    local minor=$(echo $version | cut -d. -f2)
    
    if [[ $major -lt 10 ]] || ([[ $major -eq 10 ]] && [[ $minor -lt 15 ]]); then
        print_error "需要 macOS 10.15 (Catalina) 或更高版本"
        print_error "当前版本: $version"
        exit 1
    fi
    
    print_info "检测到 macOS 版本: $version ✓"
}

# 获取最新版本
get_latest_version() {
    local version=""
    local latest_url="https://api.github.com/repos/${REPO}/releases/latest"
    local all_releases_url="https://api.github.com/repos/${REPO}/releases"
    
    # 检查下载工具
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    # 优先尝试获取最新的正式发布版本 (non-prerelease)
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -s "$latest_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -1)
    else
        version=$(wget -qO- "$latest_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -1)
    fi
    
    # 如果没有正式发布版本，尝试获取最新的预发布版本
    if [[ -z "$version" || "$version" == "null" ]]; then
        print_info "未找到正式发布版本，尝试获取预发布版本..."
        
        if command -v curl >/dev/null 2>&1; then
            version=$(curl -s "$all_releases_url" 2>/dev/null | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
        else
            version=$(wget -qO- "$all_releases_url" 2>/dev/null | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
        fi
        
        if [[ -n "$version" && "$version" != "null" ]]; then
            print_info "找到预发布版本: $version"
        fi
    fi
    
    # 如果还是获取失败，使用 latest 标签
    if [[ -z "$version" || "$version" == "null" ]]; then
        version="latest"
    fi
    
    echo "$version"
}

# 下载文件
download_file() {
    local url="$1"
    local output="$2"
    local show_error="${3:-true}"
    
    if command -v curl >/dev/null 2>&1; then
        if [[ "$show_error" == "true" ]]; then
            curl -L -o "$output" "$url"
        else
            curl -L -o "$output" "$url" 2>/dev/null
        fi
    elif command -v wget >/dev/null 2>&1; then
        if [[ "$show_error" == "true" ]]; then
            wget -O "$output" "$url"
        else
            wget -O "$output" "$url" 2>/dev/null
        fi
    else
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    return $?
}

# 验证文件校验和
verify_checksum() {
    local file="$1"
    local checksum_file="$2"
    
    if [[ -f "$checksum_file" ]]; then
        print_info "验证文件校验和..."
        if command -v md5 >/dev/null 2>&1; then
            local expected=$(cat "$checksum_file" | awk '{print $1}')
            local actual=$(md5 -q "$file")
            
            if [[ "$expected" == "$actual" ]]; then
                print_success "校验和验证通过 ✓"
                return 0
            else
                print_warning "MD5 校验和不匹配，但继续安装..."
                print_warning "预期: $expected"
                print_warning "实际: $actual"
            fi
        fi
    else
        print_warning "未找到校验和文件，跳过验证"
    fi
}

# 停止现有服务
stop_existing_service() {
    print_info "检查现有服务..."
    
    if launchctl list com.macos.guest-agent >/dev/null 2>&1; then
        print_info "停止现有服务..."
        launchctl stop com.macos.guest-agent 2>/dev/null || true
        launchctl unload /Library/LaunchDaemons/com.macos.guest-agent.plist 2>/dev/null || true
        sleep 1
    fi
}

# 安装二进制文件
install_binary() {
    local binary_file="$1"
    
    print_info "安装二进制文件到 $INSTALL_PATH"
    
    # 备份现有文件
    if [[ -f "$INSTALL_PATH" ]]; then
        print_info "备份现有文件..."
        cp "$INSTALL_PATH" "${INSTALL_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
    fi
    
    # 复制新文件
    cp "$binary_file" "$INSTALL_PATH"
    chmod +x "$INSTALL_PATH"
    
    print_success "二进制文件安装完成"
}

# 安装系统服务
install_service() {
    print_info "安装系统服务..."
    
    # 运行安装命令
    "$INSTALL_PATH" --install
    
    print_success "系统服务安装完成"
}

# 启动服务
start_service() {
    print_info "启动服务..."
    
    launchctl load /Library/LaunchDaemons/com.macos.guest-agent.plist 2>/dev/null || true
    launchctl start com.macos.guest-agent 2>/dev/null || true
    
    # 等待一下检查服务状态
    sleep 3
    
    if launchctl list com.macos.guest-agent >/dev/null 2>&1; then
        print_success "服务启动成功 ✓"
    else
        print_warning "服务可能未正常启动，请检查日志: tail -f /var/log/mac-guest-agent.log"
    fi
}

# 清理临时文件
cleanup() {
    if [[ -n "$temp_dir" && -d "$temp_dir" ]]; then
        print_info "清理临时文件..."
        cd /
        rm -rf "$temp_dir"
    fi
}

# 主安装函数
main() {
    echo -e "${BLUE}"
    echo "=============================================="
    echo "  macOS Guest Agent 自动安装脚本"
    echo "  Auto Installation Script"
    echo "=============================================="
    echo -e "${NC}"
    
    # 检查权限
    check_root
    
    # 检查系统版本
    check_macos_version
    
    # 检测架构
    local arch=$(detect_arch)
    print_info "检测到系统架构: $arch"
    
    local binary_filename=""
    
    # 根据安装模式选择不同的处理
    if [[ "$USE_LOCAL" == "true" ]]; then
        # 本地安装模式
        print_info "使用本地编译的二进制文件安装..."
        
        # 检查本地编译的二进制文件
        if [[ -f "build/mac-guest-agent" ]]; then
            binary_filename="build/mac-guest-agent"
        elif [[ -f "./mac-guest-agent" ]]; then
            binary_filename="./mac-guest-agent"
        else
            print_error "未找到本地编译的二进制文件"
            print_error "请先运行 'make build' 或确保在正确的目录中"
            exit 1
        fi
        
        print_success "找到本地二进制文件: $binary_filename"
    else
        # 在线安装模式
        # 获取最新版本
        print_info "获取最新版本信息..."
        local version=$(get_latest_version)
        if [[ "$version" == "latest" ]]; then
            print_warning "无法获取具体版本号，将使用 'latest' 标签下载"
        else
            print_info "获取到版本: $version"
        fi
        
        # 创建临时目录
        temp_dir=$(mktemp -d)
        cd "$temp_dir"
        
        # 设置清理陷阱
        trap cleanup EXIT
        
        # 构建下载 URL
        binary_filename="${BINARY_NAME}-darwin-${arch}"
        local download_url="https://github.com/${REPO}/releases/download/${version}/${binary_filename}"
        local checksum_url="https://github.com/${REPO}/releases/download/${version}/${binary_filename}.md5"
        
        print_info "下载URL: $download_url"
        
        # 下载二进制文件
        print_info "下载二进制文件..."
        if ! download_file "$download_url" "$binary_filename"; then
            print_error "下载二进制文件失败"
            print_error "请检查网络连接和版本是否存在"
            exit 1
        fi
        
        # 检查文件是否下载成功
        if [[ ! -f "$binary_filename" ]] || [[ ! -s "$binary_filename" ]]; then
            print_error "下载的文件无效或为空"
            exit 1
        fi
        
        print_success "二进制文件下载完成"
        
        # 复制configs目录（如果存在）
        if [[ -d "../configs" ]]; then
            print_info "复制 'configs' 目录到临时目录"
            cp -r ../configs .
        fi
        
        # 下载校验和文件（可选）
        if download_file "$checksum_url" "${binary_filename}.md5" false; then
            verify_checksum "$binary_filename" "${binary_filename}.md5"
        else
            print_warning "无法下载校验和文件，跳过验证"
        fi
    fi
    
    # 停止现有服务
    stop_existing_service
    
    # 安装二进制文件
    install_binary "$binary_filename"
    
    # 安装系统服务
    install_service
    
    # 启动服务
    start_service
    
    echo -e "${GREEN}"
    echo "=============================================="
    echo "  安装完成! Installation Complete!"
    echo "=============================================="
    echo -e "${NC}"
    
    print_success "macOS Guest Agent 已成功安装并启动"
    print_info "日志文件: /var/log/mac-guest-agent.log"
    print_info "配置文件: /Library/LaunchDaemons/com.macos.guest-agent.plist"
    
    echo ""
    print_info "常用命令:"
    echo "  检查服务状态: sudo launchctl list com.macos.guest-agent"
    echo "  查看日志:     tail -f /var/log/mac-guest-agent.log"
    echo "  停止服务:     sudo launchctl stop com.macos.guest-agent"
    echo "  启动服务:     sudo launchctl start com.macos.guest-agent"
    echo "  卸载服务:     sudo $INSTALL_PATH --uninstall"
    
    echo ""
    print_info "更多信息请访问: https://github.com/${REPO}"
}

# 检查参数
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "macOS Guest Agent 自动安装脚本"
    echo ""
    echo "用法: sudo $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --local         使用本地编译的二进制文件安装"
    echo "  --help, -h      显示此帮助信息"
    echo ""
    echo "此脚本将："
    echo "  1. 检测系统架构（Intel/Apple Silicon）"
    echo "  2. 下载对应的最新版本"
    echo "  3. 验证文件完整性"
    echo "  4. 安装系统服务"
    echo "  5. 启动服务"
    echo ""
    echo "系统要求:"
    echo "  - macOS 10.15+ (Catalina 或更高版本)"
    echo "  - root 权限"
    echo "  - 网络连接"
    echo ""
    exit 0
fi

# 检查是否使用本地安装
USE_LOCAL=false
if [[ "$1" == "--local" ]]; then
    USE_LOCAL=true
    shift
fi

# 运行主函数
main "$@" 