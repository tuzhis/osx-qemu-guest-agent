# 支持的命令详细说明

本文档详细介绍了 macOS Guest Agent 支持的所有 QEMU Guest Agent 命令。

## 协议参考

**QEMU Guest Agent 官方协议文档**: [QEMU Guest Agent Protocol Reference](https://www.qemu.org/docs/master/interop/qemu-ga-ref.html)

## 命令支持状态

macOS Guest Agent 目前支持 **31+** 个标准 QEMU Guest Agent 命令。

| 命令名称 | 状态 | 功能描述 | 返回数据 | 备注 |
|---------|------|----------|----------|------|
| `guest-ping` | ✅ | 心跳检查，验证guest agent是否正常运行 | 无 | 基础连接测试 |
| `guest-sync` | ✅ | 同步客户端和服务端状态 | 返回客户端传入的ID | 用于状态同步 |
| `guest-sync-id` | ✅ | 同步命令别名 | 返回客户端传入的ID | 兼容性支持 |
| `guest-sync-delimited` | ✅ | 带分隔符的同步命令 | 返回客户端传入的ID | 增强同步机制 |
| `guest-info` | ✅ | 获取客户机代理信息 | 代理版本和支持的命令列表 | 基础信息查询 |
| `guest-get-time` | ✅ | 获取当前系统时间 | 时间戳（纳秒） | 时间管理 |
| `guest-set-time` | ✅ | 设置系统时间 | 无 | 时间同步 |
| `guest-get-timezone` | ✅ | 获取系统时区信息 | 时区名称和UTC偏移 | 时区管理 |
| `guest-get-hostname` | ✅ | 获取系统主机名 | 主机名字符串 | 系统标识 |
| `guest-get-host-name` | ✅ | 获取系统主机名（别名） | 主机名字符串 | 兼容性支持 |
| `guest-get-osinfo` | ✅ | 获取操作系统详细信息 | 系统版本、内核等信息 | 系统信息 |
| `guest-get-users` | ✅ | 获取当前登录用户信息 | 用户列表和会话状态 | 用户管理 |
| `guest-get-vcpus` | ✅ | 获取虚拟CPU信息 | CPU核心数和状态 | 硬件信息 |
| `guest-get-memory-blocks` | ✅ | 获取内存块列表 | 内存块详细信息 | 内存管理 |
| `guest-get-memory-block-info` | ✅ | 获取内存块配置信息 | 内存块大小等信息 | 内存配置 |
| `guest-set-memory-blocks` | ✅ | 设置内存块状态 | 无 | 内存热插拔（模拟） |
| `guest-get-memory-info` | ✅ | 获取详细内存使用情况 | 内存统计信息 | macOS特有扩展 |
| `guest-network-get-interfaces` | ✅ | 获取网络接口信息 | 网络接口列表和配置 | 网络管理 |
| `guest-get-fsinfo` | ✅ | 获取文件系统信息 | 文件系统挂载点和类型 | 存储信息 |
| `guest-get-disks` | ✅ | 获取磁盘信息 | 磁盘列表和分区信息 | 存储管理 |
| `guest-fsfreeze-status` | ✅ | 获取文件系统冻结状态 | 冻结状态（thawed/frozen） | 文件系统管理 |
| `guest-fsfreeze-freeze` | ✅ | 冻结文件系统 | 冻结的文件系统数量 | 快照支持 |
| `guest-fsfreeze-thaw` | ✅ | 解冻文件系统 | 解冻的文件系统数量 | 快照支持 |
| `guest-fstrim` | ✅ | 执行文件系统trim操作 | Trim操作结果 | 存储优化 |
| `guest-shutdown` | ✅ | 关机、重启或挂起系统 | 无返回（异步操作） | 电源管理 |
| `guest-suspend-disk` | ✅ | 挂起到磁盘（休眠） | 无返回（异步操作） | 电源管理 |
| `guest-suspend-ram` | ✅ | 挂起到内存（睡眠） | 无返回（异步操作） | 电源管理 |
| `guest-suspend-hybrid` | ✅ | 混合挂起模式 | 无返回（异步操作） | 电源管理 |
| `guest-ssh-get-authorized-keys` | ⚠️ | 获取SSH授权密钥 | 密钥列表 | 安全限制，仅记录请求 |
| `guest-ssh-add-authorized-keys` | ⚠️ | 添加SSH授权密钥 | 无 | 安全限制，仅记录请求 |
| `guest-ssh-remove-authorized-keys` | ⚠️ | 移除SSH授权密钥 | 无 | 安全限制，仅记录请求 |
| `guest-exec` | ⚠️ | 在客户机中执行命令 | 进程ID | 安全限制，仅记录请求 |
| `guest-exec-status` | ⚠️ | 获取执行命令的状态 | 进程状态信息 | 安全限制，仅记录请求 |

## 命令分类详细说明

### 🔧 基础命令

#### `guest-ping`
- **功能**: 验证guest agent连接状态
- **参数**: 无
- **返回**: 成功则无错误
- **用途**: 健康检查和连接测试

#### `guest-sync` / `guest-sync-delimited` / `guest-sync-id`
- **功能**: 同步客户端和服务端状态
- **参数**: `id` (整数) - 随机生成的64位整数
- **返回**: 返回客户端传入的ID
- **用途**: 确保通信流同步，避免脏数据
- **备注**: `guest-sync-id`是`guest-sync`的别名，提供兼容性支持

#### `guest-info`
- **功能**: 获取agent信息和支持的命令列表
- **参数**: 无
- **返回**: `GuestAgentInfo` 对象
- **用途**: 功能发现和版本检查

### 📊 系统信息类

#### `guest-get-osinfo`
- **功能**: 获取详细的操作系统信息
- **参数**: 无
- **返回**: `GuestOSInfo` 对象，包含：
  - `kernel-release`: 内核版本
  - `machine`: 机器架构
  - `name`: 操作系统名称
  - `pretty-name`: 友好显示名称
  - `version`: 系统版本

#### `guest-get-hostname` / `guest-get-host-name`
- **功能**: 获取系统主机名
- **参数**: 无
- **返回**: `GuestHostName` 对象
- **用途**: 系统标识和网络配置
- **备注**: 两个命令名称都受支持，提供最大兼容性

#### `guest-get-time` / `guest-set-time`
- **功能**: 时间管理操作
- **参数**: 
  - `guest-get-time`: 无
  - `guest-set-time`: `time` (可选) - 纳秒时间戳
- **返回**: 
  - `guest-get-time`: 当前时间戳
  - `guest-set-time`: 无
- **用途**: 时间同步和管理

#### `guest-get-timezone`
- **功能**: 获取时区信息
- **参数**: 无
- **返回**: `GuestTimezone` 对象
- **用途**: 时区配置和本地化

### 👥 用户管理

#### `guest-get-users`
- **功能**: 获取当前活跃用户会话
- **参数**: 无
- **返回**: `GuestUser` 数组，包含：
  - `user`: 用户名
  - `login-time`: 登录时间
  - `domain`: 登录域（Windows）
- **用途**: 用户会话监控

### 🖥️ 硬件信息

#### `guest-get-vcpus`
- **功能**: 获取虚拟CPU信息
- **参数**: 无
- **返回**: `GuestLogicalProcessor` 数组
- **用途**: CPU状态监控和配置

#### `guest-get-memory-blocks` / `guest-get-memory-block-info`
- **功能**: 内存管理信息
- **参数**: 无
- **返回**: 内存块详细信息
- **用途**: 内存热插拔和管理

#### `guest-set-memory-blocks`
- **功能**: 设置内存块状态（上线/下线）
- **参数**: 内存块索引和目标状态
- **返回**: 无
- **备注**: 在macOS上为模拟实现，实际不支持内存热插拔

#### `guest-get-memory-info`
- **功能**: 获取详细内存使用情况
- **参数**: 无
- **返回**: 内存统计信息（包括活跃/不活跃/空闲/已用内存等）
- **备注**: macOS特有的扩展命令，提供更详细的内存使用情况

### 🌐 网络管理

#### `guest-network-get-interfaces`
- **功能**: 获取网络接口详细信息
- **参数**: 无
- **返回**: `GuestNetworkInterface` 数组，包含：
  - `name`: 接口名称
  - `hardware-address`: MAC地址
  - `ip-addresses`: IP地址列表
  - `statistics`: 网络统计信息
- **用途**: 网络配置和监控

### 💾 文件系统操作

#### `guest-get-fsinfo`
- **功能**: 获取文件系统挂载信息
- **参数**: 无
- **返回**: `GuestFilesystemInfo` 数组
- **用途**: 存储管理和监控

#### `guest-get-disks`
- **功能**: 获取磁盘和分区信息
- **参数**: 无
- **返回**: `GuestDiskInfo` 数组，包含：
  - `name`: 磁盘名称
  - `partition`: 是否为分区
  - `size`: 磁盘大小
  - `partitions`: 分区列表
- **用途**: 磁盘管理和监控

#### 文件系统冻结操作
- **`guest-fsfreeze-status`**: 查询冻结状态
- **`guest-fsfreeze-freeze`**: 冻结所有文件系统
- **`guest-fsfreeze-thaw`**: 解冻所有文件系统
- **用途**: 快照和备份前的文件系统一致性保证

#### `guest-fstrim`
- **功能**: 对文件系统执行TRIM操作
- **参数**: `minimum` (可选) - 最小连续空闲范围
- **返回**: `GuestFilesystemTrimResponse` 对象
- **用途**: SSD优化和空间回收

### ⚡ 电源管理

#### `guest-shutdown`
- **功能**: 系统关机/重启操作
- **参数**: `mode` (可选) - "halt", "powerdown", "reboot"
- **返回**: 无（异步操作）
- **用途**: 系统电源控制

#### 挂起操作
- **`guest-suspend-disk`**: 挂起到磁盘（休眠）
- **`guest-suspend-ram`**: 挂起到内存（睡眠）
- **`guest-suspend-hybrid`**: 混合挂起模式
- **用途**: 节能和快速恢复

### 🔒 命令执行（安全限制）

#### `guest-exec` / `guest-exec-status`
- **功能**: 在客户机中执行命令并获取状态
- **参数**: 
  - `guest-exec`: 命令路径、参数、环境变量等
  - `guest-exec-status`: 进程ID
- **返回**: 
  - `guest-exec`: 进程ID
  - `guest-exec-status`: 进程状态信息
- **安全说明**: 出于安全考虑，macOS Guest Agent 不实际执行命令，仅记录请求并返回友好错误信息
- **用途**: 自动化脚本和远程管理

#### `guest-ssh-get-authorized-keys` / `guest-ssh-add-authorized-keys` / `guest-ssh-remove-authorized-keys`
- **功能**: 管理SSH授权密钥
- **参数**:
  - `guest-ssh-get-authorized-keys`: 用户名
  - `guest-ssh-add-authorized-keys`: 用户名和SSH密钥列表
  - `guest-ssh-remove-authorized-keys`: 用户名和SSH密钥列表
- **返回**:
  - `guest-ssh-get-authorized-keys`: 授权密钥列表
  - 其他命令: 无
- **安全说明**: 出于安全考虑，macOS Guest Agent 不实际管理SSH密钥，仅记录请求并返回友好错误信息
- **用途**: 自动化SSH密钥管理和远程访问控制

## 使用示例

### 基础连接测试
```json
-> {"execute": "guest-ping"}
<- {"return": {}}
```

### 获取系统信息
```json
-> {"execute": "guest-get-osinfo"}
<- {"return": {"name": "macOS", "version": "14.0", "machine": "arm64"}}
```

### 网络接口查询
```json
-> {"execute": "guest-network-get-interfaces"}
<- {"return": [{"name": "en0", "hardware-address": "xx:xx:xx:xx:xx:xx"}]}
```

### 文件系统操作
```json
-> {"execute": "guest-fsfreeze-freeze"}
<- {"return": 3}

-> {"execute": "guest-fsfreeze-thaw"}
<- {"return": 3}
```

## 平台特性

### macOS 特定实现
- **自动QEMU环境检测**: 仅在QEMU虚拟化环境中运行
- **多种挂起策略**: 支持多种备用挂起方法
- **文件系统优化**: 特针对APFS和HFS+优化
- **网络统计**: 完整的网络接口统计信息
- **命令别名支持**: 提供多种命令名称变体以增强兼容性

### 兼容性说明
- 所有命令遵循QEMU Guest Agent标准协议
- 支持标准QMP (QEMU Machine Protocol) 格式
- 与主流虚拟化平台兼容（QEMU/KVM, PVE等）
- 特别优化了与PVE (Proxmox Virtual Environment) 的兼容性

## 技术参考

- **QEMU Guest Agent 规范**: [官方文档](https://www.qemu.org/docs/master/interop/qemu-ga-ref.html)
- **QMP 协议规范**: [QMP 文档](https://www.qemu.org/docs/master/interop/qmp-spec.html)
- **源码实现**: 本项目 `/internal/commands/` 目录 