#!/bin/bash

# macOS Guest Agent 卸载脚本
set -e

PROGRAM_NAME="mac-guest-agent"
BINARY_PATH="/usr/local/bin/${PROGRAM_NAME}"
PLIST_PATH="/Library/LaunchDaemons/com.macos.guest-agent.plist"
SHARE_PATH="/usr/local/share/${PROGRAM_NAME}"
LOG_PATH="/var/log/${PROGRAM_NAME}.log"

echo "开始卸载 macOS Guest Agent..."

# 检查是否以root权限运行
if [[ $EUID -ne 0 ]]; then
   echo "错误: 此脚本需要root权限运行" 
   echo "请使用: sudo $0"
   exit 1
fi

# 停止并卸载服务
if launchctl list | grep -q "com.macos.guest-agent"; then
    echo "停止 Guest Agent 服务..."
    launchctl stop com.macos.guest-agent 2>/dev/null || true
    launchctl unload "$PLIST_PATH" 2>/dev/null || true
    echo "服务已停止"
else
    echo "服务未运行"
fi

# 删除文件
echo "删除程序文件..."

if [[ -f "$BINARY_PATH" ]]; then
    rm -f "$BINARY_PATH"
    echo "已删除: $BINARY_PATH"
fi

if [[ -f "$PLIST_PATH" ]]; then
    rm -f "$PLIST_PATH"
    echo "已删除: $PLIST_PATH"
fi

if [[ -d "$SHARE_PATH" ]]; then
    rm -rf "$SHARE_PATH"
    echo "已删除: $SHARE_PATH"
fi

# 询问是否删除日志文件
if [[ -f "$LOG_PATH" ]]; then
    echo ""
    read -p "是否删除日志文件 $LOG_PATH? [y/N]: " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f "$LOG_PATH"
        echo "已删除日志文件"
    else
        echo "保留日志文件"
    fi
fi

echo ""
echo "卸载完成!" 