# macOS Guest Agent / macOS 虚拟机代理

[中文](#中文) | [English](#english)

---

## 中文

### 概述

macOS Guest Agent 是一个为 macOS 虚拟机设计的轻量级 QEMU Guest Agent 实现。它在宿主系统和 macOS 客户机之间提供通信通道，通过 QMP（QEMU Machine Protocol）接口实现各种系统操作和信息查询。

**使用场景**: 专门解决 PVE 等虚拟化平台中 macOS 虚拟机无法正确响应关机、重启等电源管理指令的问题，让宿主机能够通过标准 QEMU Guest Agent 命令有效控制 macOS 虚拟机。

### 功能特性

- **系统信息**：获取操作系统信息、主机名、时间、时区、硬件详情
- **电源管理**：关机、重启、挂起操作，支持多种备用方法
- **文件系统操作**：冻结/解冻文件系统、trim 操作、文件系统信息
- **网络信息**：网络接口详情和配置
- **用户管理**：当前用户会话和信息
- **内存和CPU**：内存块信息和虚拟CPU详情
- **存储管理**：磁盘和分区信息
- **QEMU环境检测**：自动检测是否运行在 QEMU 虚拟化环境中
- **系统服务**：作为 macOS LaunchDaemon 运行，支持自动启动
- **多种通信方式**：支持 virtio-serial 通信
- **命令兼容性**：支持多种命令别名，增强兼容性
- **安全限制**：对敏感命令（如SSH密钥管理、命令执行）实施安全限制

### 支持的命令

macOS Guest Agent 支持 **31+** 种标准 QEMU Guest Agent 命令。

**📖 查看完整命令列表和详细说明**: [支持的命令详细说明](wiki/COMMANDS.md)

### 快速安装

#### 一键安装（推荐）

```bash
# 自动检测架构并安装最新版本
curl -fsSL https://raw.githubusercontent.com/tuzhis/osx-qemu-guest-agent/main/scripts/install.sh | sudo bash
```

#### 手动安装

```bash
# 从 GitHub Releases 下载对应架构的二进制文件
# AMD64 架构
wget https://github.com/tuzhis/osx-qemu-guest-agent/releases/latest/download/mac-guest-agent-darwin-amd64

# ARM64 架构  
wget https://github.com/tuzhis/osx-qemu-guest-agent/releases/latest/download/mac-guest-agent-darwin-arm64

# 添加执行权限并安装
chmod +x mac-guest-agent-darwin-*
sudo ./mac-guest-agent-darwin-* --install
```

#### 从源码构建

```bash
git clone https://github.com/tuzhis/osx-qemu-guest-agent.git
cd osx-qemu-guest-agent
sudo make install
```

### 系统要求

- macOS 10.15 或更高版本
- 安装系统服务需要 root 权限
- 从源码构建需要 Go 1.21+

### 使用方法

#### 系统服务模式

安装后代理会自动运行：

```bash
# 检查服务状态
sudo launchctl list com.macos.guest-agent

# 查看日志
tail -f /var/log/mac-guest-agent.log

# 管理服务
sudo launchctl stop com.macos.guest-agent
sudo launchctl start com.macos.guest-agent

# 卸载服务
sudo /usr/local/bin/mac-guest-agent --uninstall
```



### PVE 环境验证

安装完成后，可以直接在PVE宿主机上验证功能。

#### 手动验证基础命令

```bash
# 在PVE宿主机执行标准命令（替换100为实际VM ID）
qm guest ping 100        # 心跳测试
qm guest info 100        # 获取代理信息
qm guest cmd 100 get-osinfo     # 获取系统信息
qm guest cmd 100 get-hostname   # 获取主机名
```

#### 快速测试所有无风险指令

使用以下一行命令下载并运行完整测试脚本：

```bash
# 快速测试所有无风险的查询命令
curl -fsSL https://raw.githubusercontent.com/tuzhis/osx-qemu-guest-agent/main/scripts/pve_qemu_agent_test.sh | bash
```

**注意**: 脚本会提示输入VM ID，只测试无风险的查询命令，跳过可能影响系统的操作。

### 问题排查

常见问题解决方案：

1. **权限拒绝**：确保使用 root 权限运行
2. **设备未找到**：确认运行在启用了 guest agent 的 QEMU 环境中
3. **服务无法启动**：检查 `/var/log/mac-guest-agent.log` 中的日志

### 开发说明

```bash
# 安装依赖
go mod download

# 构建
make build

# 运行测试
make test

# 清理构建产物
make clean
```

---

## English

### Overview

macOS Guest Agent is a lightweight QEMU Guest Agent implementation for macOS virtual machines. It provides a communication channel between the host system and macOS guest, enabling various system operations and information queries through the QMP (QEMU Machine Protocol) interface.

**Use Case**: Specifically addresses the issue where macOS virtual machines in PVE and other virtualization platforms cannot properly respond to shutdown, reboot, and other power management commands, enabling host systems to effectively control macOS VMs through standard QEMU Guest Agent commands.

### Features

- **System Information**: Get OS info, hostname, time, timezone, hardware details
- **Power Management**: Shutdown, reboot, suspend operations with multiple fallback methods
- **File System Operations**: Freeze/thaw file systems, trim operations, file system info
- **Network Information**: Network interface details and configuration
- **User Management**: Current user sessions and information
- **Memory & CPU**: Memory block info and virtual CPU details
- **Storage Management**: Disk and partition information
- **QEMU Environment Detection**: Automatically detects if running in QEMU virtualization
- **System Service**: Runs as macOS LaunchDaemon with automatic startup
- **Multiple Communication Methods**: Supports virtio-serial communication
- **Command Compatibility**: Supports multiple command aliases for enhanced compatibility
- **Security Restrictions**: Implements security restrictions for sensitive commands (SSH key management, command execution)

### Supported Commands

macOS Guest Agent supports **31+** standard QEMU Guest Agent commands.

**📖 View complete command list and detailed documentation**: [Supported Commands Documentation](wiki/COMMANDS.md)

### Quick Installation

#### One-line Install (Recommended)

```bash
# Auto-detect architecture and install latest version
curl -fsSL https://raw.githubusercontent.com/tuzhis/osx-qemu-guest-agent/main/scripts/install.sh | sudo bash
```

#### Manual Installation

```bash
# Download architecture-specific binary from GitHub Releases
# For AMD64
wget https://github.com/tuzhis/osx-qemu-guest-agent/releases/latest/download/mac-guest-agent-darwin-amd64

# For ARM64
wget https://github.com/tuzhis/osx-qemu-guest-agent/releases/latest/download/mac-guest-agent-darwin-arm64

# Add execute permission and install
chmod +x mac-guest-agent-darwin-*
sudo ./mac-guest-agent-darwin-* --install
```

#### Build from Source

```bash
git clone https://github.com/tuzhis/osx-qemu-guest-agent.git
cd osx-qemu-guest-agent
sudo make install
```

### Prerequisites

- macOS 10.15 or later
- Root privileges for system service installation
- Go 1.21+ for building from source

### Usage

#### System Service Mode

Once installed, the agent runs automatically:

```bash
# Check service status
sudo launchctl list com.macos.guest-agent

# View logs
tail -f /var/log/mac-guest-agent.log

# Manage service
sudo launchctl stop com.macos.guest-agent
sudo launchctl start com.macos.guest-agent

# Uninstall service
sudo /usr/local/bin/mac-guest-agent --uninstall
```



### PVE Environment Verification

After installation, you can directly verify functionality on the PVE host.

#### Manual Verification with Basic Commands

```bash
# Execute standard commands from PVE host (replace 100 with actual VM ID)
qm guest ping 100        # Heartbeat test
qm guest info 100        # Get agent information
qm guest cmd 100 get-osinfo     # Get system information
qm guest cmd 100 get-hostname   # Get hostname
```

#### Quick Test of All Safe Commands

Use the following one-line command to download and run the comprehensive test script:

```bash
# Quick test of all safe query commands
curl -fsSL https://raw.githubusercontent.com/tuzhis/osx-qemu-guest-agent/main/scripts/pve_qemu_agent_test.sh | bash
```

**Note**: The script will prompt for VM ID input and only tests safe query commands, skipping operations that might affect the system.

### Troubleshooting

Common issues and solutions:

1. **Permission Denied**: Ensure running with root privileges
2. **Device Not Found**: Verify running in QEMU environment with guest agent enabled
3. **Service Won't Start**: Check logs at `/var/log/mac-guest-agent.log`

### Development

```bash
# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Project Structure

```
osx-qemu-guest-agent/
├── cmd/main.go              # Main application entry
├── internal/
│   ├── agent/               # Core agent logic
│   ├── commands/            # Command handlers
│   ├── communication/       # Device communication
│   └── protocol/            # QMP protocol handling
├── configs/                 # LaunchDaemon configuration
├── scripts/                 # Build and installation scripts
└── pve_qemu_agent_test.sh  # PVE testing script
```

---

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 