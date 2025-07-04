#!/bin/bash

# macOS Guest Agent 本地安装脚本
# 这个脚本简化了安装过程，直接使用本地编译的二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
if [[ $EUID -ne 0 ]]; then
    print_error "此脚本需要 root 权限运行。请使用 sudo 执行："
    print_error "sudo $0"
    exit 1
fi

echo -e "${BLUE}"
echo "=============================================="
echo "  macOS Guest Agent 本地安装脚本"
echo "  Local Installation Script"
echo "=============================================="
echo -e "${NC}"

# 检查本地二进制文件
if [[ ! -f "build/mac-guest-agent" ]]; then
    print_error "未找到本地编译的二进制文件"
    print_error "请先运行 'make build' 构建项目"
    exit 1
fi

print_success "找到本地二进制文件: build/mac-guest-agent"

# 安装二进制文件
print_info "安装二进制文件到 /usr/local/bin/mac-guest-agent"
cp "build/mac-guest-agent" /usr/local/bin/mac-guest-agent
chmod +x /usr/local/bin/mac-guest-agent
print_success "二进制文件安装完成"

# 安装系统服务
print_info "安装系统服务..."
/usr/local/bin/mac-guest-agent --install
print_success "系统服务安装完成"

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
echo "  卸载服务:     sudo /usr/local/bin/mac-guest-agent --uninstall" 