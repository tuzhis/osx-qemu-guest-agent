#!/bin/bash

# PVE QEMU Guest Agent 快速测试脚本
# 用于测试 macOS Guest Agent 的所有无风险功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 获取脚本名称
SCRIPT_NAME=$(basename "$0")

# 显示帮助信息
show_help() {
    echo -e "${BLUE}PVE QEMU Guest Agent 测试脚本${NC}"
    echo -e "${BLUE}用于测试 macOS Guest Agent 的所有无风险功能${NC}"
    echo ""
    echo -e "${YELLOW}用法：${NC}"
    echo "  $SCRIPT_NAME [VM_ID]"
    echo ""
    echo -e "${YELLOW}参数：${NC}"
    echo "  VM_ID    虚拟机ID（可选，如不提供会提示输入）"
    echo ""
    echo -e "${YELLOW}示例：${NC}"
    echo "  $SCRIPT_NAME 100      # 测试VM 100"
    echo "  $SCRIPT_NAME          # 交互式输入VM ID"
    echo ""
    echo -e "${YELLOW}注意：${NC}"
    echo "  - 此脚本只执行无风险的查询命令"
    echo "  - 跳过可能影响系统的操作（关机、重启、文件系统冻结等）"
    echo "  - 需要在PVE宿主机上运行，且具有qm命令权限"
}

# 检查参数
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

if [ $# -eq 0 ]; then
    echo -e "${CYAN}请输入虚拟机ID:${NC}"
    read -p "VM ID: " VMID
    if [ -z "$VMID" ]; then
        echo -e "${RED}错误: 虚拟机ID不能为空${NC}"
        exit 1
    fi
elif [ $# -eq 1 ]; then
    VMID=$1
else
    echo -e "${RED}错误: 参数过多${NC}"
    echo ""
    show_help
    exit 1
fi

# 验证VM ID是数字
if ! [[ "$VMID" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}错误: 虚拟机ID必须是数字${NC}"
    exit 1
fi

# 检查是否有qm命令
if ! command -v qm >/dev/null 2>&1; then
    echo -e "${RED}错误: 未找到qm命令，请确保在PVE宿主机上运行${NC}"
    exit 1
fi

# 检查虚拟机是否存在
if ! qm status $VMID >/dev/null 2>&1; then
    echo -e "${RED}错误: 虚拟机 $VMID 不存在${NC}"
    exit 1
fi

# 检查虚拟机是否运行
VM_STATUS=$(qm status $VMID | grep -o "status: [a-z]*" | cut -d' ' -f2)
if [ "$VM_STATUS" != "running" ]; then
    echo -e "${RED}错误: 虚拟机 $VMID 状态为 $VM_STATUS，需要运行状态${NC}"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PVE QEMU Guest Agent 测试脚本${NC}"
echo -e "${BLUE}  虚拟机ID: $VMID${NC}"
echo -e "${BLUE}  虚拟机状态: ${GREEN}$VM_STATUS${NC}"
echo -e "${BLUE}========================================${NC}"
echo

# 测试函数
run_test() {
    local test_name="$1"
    local command="$2"
    local description="$3"
    
    echo -e "${CYAN}测试: $test_name${NC}"
    echo -e "描述: $description"
    echo -n "执行: $command..."
    
    # 创建临时文件保存输出
    local temp_output=$(mktemp)
    local temp_error=$(mktemp)
    
    if timeout 15 $command > "$temp_output" 2> "$temp_error"; then
        echo -e " ${GREEN}✓ 成功${NC}"
        if [ -s "$temp_output" ]; then
            echo -e "${GREEN}输出:${NC}"
            sed 's/^/  /' "$temp_output"
        fi
        rm -f "$temp_output" "$temp_error"
        echo
        return 0
    else
        echo -e " ${RED}✗ 失败${NC}"
        if [ -s "$temp_error" ]; then
            echo -e "${RED}错误:${NC}"
            sed 's/^/  /' "$temp_error"
        fi
        rm -f "$temp_output" "$temp_error"
        echo
        return 1
    fi
}

# 测试计数器
total_tests=0
passed_tests=0

# 基础连接测试
echo -e "${BLUE}=== 基础连接测试 ===${NC}"
if run_test "心跳测试" "qm guest ping $VMID" "测试与guest agent的基本连接"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "代理信息" "qm guest info $VMID" "获取guest agent版本和支持的命令"; then
    ((passed_tests++))
fi
((total_tests++))

# 系统信息测试
echo -e "${BLUE}=== 系统信息测试 ===${NC}"
if run_test "操作系统信息" "qm guest cmd $VMID get-osinfo" "获取操作系统详细信息"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "主机名获取" "qm guest cmd $VMID get-hostname" "获取系统主机名"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "系统时间" "qm guest cmd $VMID get-time" "获取系统当前时间"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "时区信息" "qm guest cmd $VMID get-timezone" "获取系统时区设置"; then
    ((passed_tests++))
fi
((total_tests++))

# 用户和进程测试
echo -e "${BLUE}=== 用户和进程测试 ===${NC}"
if run_test "用户会话" "qm guest cmd $VMID get-users" "获取当前登录用户信息"; then
    ((passed_tests++))
fi
((total_tests++))

# 硬件信息测试
echo -e "${BLUE}=== 硬件信息测试 ===${NC}"
if run_test "虚拟CPU信息" "qm guest cmd $VMID get-vcpus" "获取虚拟CPU配置信息"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "内存块信息" "qm guest cmd $VMID get-memory-block-info" "获取内存块配置信息"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "内存块列表" "qm guest cmd $VMID get-memory-blocks" "获取内存块详细列表"; then
    ((passed_tests++))
fi
((total_tests++))

# 网络信息测试
echo -e "${BLUE}=== 网络信息测试 ===${NC}"
if run_test "网络接口" "qm guest cmd $VMID network-get-interfaces" "获取网络接口配置信息"; then
    ((passed_tests++))
fi
((total_tests++))

# 文件系统测试（仅查询命令）
echo -e "${BLUE}=== 文件系统测试 ===${NC}"
if run_test "文件系统信息" "qm guest cmd $VMID get-fsinfo" "获取文件系统挂载信息"; then
    ((passed_tests++))
fi
((total_tests++))

if run_test "文件系统冻结状态" "qm guest cmd $VMID fsfreeze-status" "检查文件系统冻结状态"; then
    ((passed_tests++))
fi
((total_tests++))

# 显示跳过的测试及手动执行方法
echo -e "${YELLOW}=== 跳过的高风险测试 ===${NC}"
echo -e "${YELLOW}以下命令可能影响系统稳定性，需要手动执行并承担风险：${NC}"
echo
echo -e "${RED}⚠️  文件系统冻结/解冻操作：${NC}"
echo -e "   ${CYAN}qm guest cmd $VMID fsfreeze-freeze${NC}    # 冻结所有文件系统"
echo -e "   ${CYAN}qm guest cmd $VMID fsfreeze-thaw${NC}      # 解冻所有文件系统"
echo -e "   ${YELLOW}说明: 用于创建一致性快照，冻结期间磁盘I/O会暂停${NC}"
echo
echo -e "${RED}⚠️  电源管理操作：${NC}"
echo -e "   ${CYAN}qm guest cmd $VMID shutdown${NC}           # 优雅关机"
echo -e "   ${CYAN}qm guest cmd $VMID shutdown '{\"mode\":\"halt\"}'${NC}  # 停机不关电源"
echo -e "   ${CYAN}qm guest cmd $VMID shutdown '{\"mode\":\"reboot\"}'${NC} # 重启系统"
echo -e "   ${YELLOW}说明: 会导致系统关机或重启，请确保保存所有工作${NC}"
echo
echo -e "${RED}⚠️  挂起操作：${NC}"
echo -e "   ${CYAN}qm guest cmd $VMID suspend-disk${NC}       # 挂起到磁盘"
echo -e "   ${CYAN}qm guest cmd $VMID suspend-ram${NC}        # 挂起到内存"
echo -e "   ${CYAN}qm guest cmd $VMID suspend-hybrid${NC}     # 混合挂起"
echo -e "   ${YELLOW}说明: 会使系统进入睡眠或休眠状态${NC}"
echo
echo -e "${RED}⚠️  时间设置：${NC}"
echo -e "   ${CYAN}qm guest cmd $VMID set-time '{\"time\":\$(date +%s)000000000}'${NC}"
echo -e "   ${YELLOW}说明: 设置系统时间，可能影响时间敏感的应用${NC}"
echo
echo -e "${RED}⚠️  文件系统TRIM：${NC}"
echo -e "   ${CYAN}qm guest cmd $VMID fstrim '{\"minimum\":0}'${NC}     # 对所有文件系统执行TRIM"
echo -e "   ${CYAN}qm guest cmd $VMID fstrim '{\"minimum\":0,\"mountpoint\":\"/\"}'${NC}  # 指定挂载点"
echo -e "   ${YELLOW}说明: 优化SSD性能，执行时间较长，可能暂时影响磁盘性能${NC}"
echo
echo -e "${RED}❗ 重要提醒：${NC}"
echo -e "${RED}   • 执行前请确保重要数据已备份${NC}"
echo -e "${RED}   • 在生产环境中请谨慎使用${NC}"
echo -e "${RED}   • 文件系统冻结/解冻操作必须成对使用${NC}"
echo -e "${RED}   • 建议先在测试环境中验证${NC}"
echo

# 显示测试结果
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  测试结果统计${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "总测试数: $total_tests"
echo -e "通过测试: ${GREEN}$passed_tests${NC}"
echo -e "失败测试: ${RED}$((total_tests - passed_tests))${NC}"

if [ $passed_tests -eq $total_tests ]; then
    echo -e "${GREEN}🎉 所有测试通过！macOS Guest Agent 工作正常。${NC}"
    echo -e "${GREEN}   所有无风险的查询功能都能正常响应。${NC}"
    exit 0
else
    success_rate=$((passed_tests * 100 / total_tests))
    echo -e "${YELLOW}⚠️  成功率: $success_rate%${NC}"
    if [ $success_rate -ge 80 ]; then
        echo -e "${YELLOW}大部分功能正常，请检查失败的测试项。${NC}"
        echo -e "${YELLOW}建议查看guest agent日志获取更多信息。${NC}"
        exit 0
    else
        echo -e "${RED}❌ 多个测试失败，请检查以下项目：${NC}"
        echo -e "${RED}   1. guest agent服务是否正常运行${NC}"
        echo -e "${RED}   2. virtio-serial设备是否正确配置${NC}"
        echo -e "${RED}   3. 查看guest agent日志文件${NC}"
        exit 1
    fi
fi 