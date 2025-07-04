#!/bin/bash

# PVE宿主机测试macOS Guest Agent脚本
# 使用方法: ./pve-test.sh VM_ID
# 例如: ./pve-test.sh 100

set -e

VM_ID=${1}
QGA_SOCKET="/var/run/qemu-server/${VM_ID}.qga"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [[ -z "$VM_ID" ]]; then
    echo -e "${RED}错误: 请提供VM ID${NC}"
    echo "使用方法: $0 <VM_ID>"
    echo "例如: $0 100"
    exit 1
fi

# 检查是否为root用户
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}错误: 此脚本需要root权限运行${NC}"
   echo "请使用: sudo $0 $VM_ID"
   exit 1
fi

echo -e "${YELLOW}==========================================${NC}"
echo -e "${YELLOW}测试macOS Guest Agent - VM ID: ${VM_ID}${NC}"
echo -e "${YELLOW}==========================================${NC}"

# 检查VM是否存在
if [[ ! -f "/etc/pve/qemu-server/${VM_ID}.conf" ]]; then
    echo -e "${RED}❌ VM ${VM_ID} 不存在${NC}"
    exit 1
fi

# 检查VM是否运行
VM_STATUS=$(qm status ${VM_ID} | awk '{print $2}')
if [[ "$VM_STATUS" != "running" ]]; then
    echo -e "${RED}❌ VM ${VM_ID} 未运行 (状态: ${VM_STATUS})${NC}"
    echo "请先启动虚拟机: qm start ${VM_ID}"
    exit 1
fi

echo -e "${GREEN}✅ VM ${VM_ID} 正在运行${NC}"

# 检查VM配置中是否启用了Guest Agent
AGENT_CONFIG=$(qm config ${VM_ID} | grep "^agent:" | cut -d: -f2 | xargs)
if [[ "$AGENT_CONFIG" != "1" ]]; then
    echo -e "${YELLOW}⚠️  VM配置中Guest Agent未启用${NC}"
    echo "当前配置: agent: ${AGENT_CONFIG:-未设置}"
    echo "请运行: qm set ${VM_ID} --agent 1"
    echo "然后重启VM: qm shutdown ${VM_ID} && qm start ${VM_ID}"
    exit 1
fi

echo -e "${GREEN}✅ VM配置中Guest Agent已启用${NC}"

# 检查socket文件
if [[ ! -S "$QGA_SOCKET" ]]; then
    echo -e "${RED}❌ Guest Agent socket不存在: $QGA_SOCKET${NC}"
    echo "可能的原因:"
    echo "1. Guest Agent未在macOS中安装或启动"
    echo "2. 需要等待更长时间让Guest Agent初始化"
    echo "3. macOS防火墙阻止了通信"
    exit 1
fi

echo -e "${GREEN}✅ Guest Agent socket存在: $QGA_SOCKET${NC}"

# 测试函数
test_command() {
    local cmd="$1"
    local desc="$2"
    
    echo -e -n "${YELLOW}测试 $desc ... ${NC}"
    
    # 使用timeout避免永久等待
    response=$(timeout 10s bash -c "echo '$cmd' | socat - unix:$QGA_SOCKET" 2>/dev/null)
    exit_code=$?
    
    if [[ $exit_code -eq 0 && -n "$response" ]]; then
        echo -e "${GREEN}✅ 成功${NC}"
        echo -e "   ${GREEN}响应: $response${NC}"
    elif [[ $exit_code -eq 124 ]]; then
        echo -e "${RED}❌ 超时${NC}"
        echo -e "   ${RED}Guest Agent可能未响应${NC}"
    else
        echo -e "${RED}❌ 失败${NC}"
        echo -e "   ${RED}无响应或连接错误${NC}"
    fi
    echo
}

# 执行基础测试
echo -e "${YELLOW}开始基础功能测试...${NC}"
echo

test_command '{"execute":"ping"}' "心跳检测 (ping)"
test_command '{"execute":"info"}' "系统信息 (info)"
test_command '{"execute":"sync","arguments":{"id":12345}}' "协议同步 (sync)"

echo -e "${YELLOW}=========================================${NC}"
echo -e "${GREEN}基础测试完成！${NC}"
echo

# 高级测试选项
echo -e "${YELLOW}高级测试选项:${NC}"
echo "1. 测试获取网络信息"
echo "2. 测试文件系统信息"
echo "3. 测试关机功能 (⚠️ 危险)"
echo "4. 退出"
echo

read -p "请选择测试项目 (1-4): " choice

case $choice in
    1)
        echo -e "${YELLOW}测试网络信息...${NC}"
        test_command '{"execute":"guest-network-get-interfaces"}' "网络接口信息"
        ;;
    2)
        echo -e "${YELLOW}测试文件系统信息...${NC}"
        test_command '{"execute":"guest-get-fsinfo"}' "文件系统信息"
        ;;
    3)
        echo -e "${RED}⚠️  警告: 这将关闭虚拟机！${NC}"
        echo -e "${RED}请确保已保存所有重要工作！${NC}"
        read -p "确认要测试关机功能吗? (输入 'YES' 确认): " confirm
                 if [[ "$confirm" == "YES" ]]; then
             echo -e "${YELLOW}执行关机测试...${NC}"
             test_command '{"execute":"shutdown","arguments":{"mode":"powerdown"}}' "关机测试"
         else
            echo "关机测试已取消"
        fi
        ;;
    4)
        echo "退出测试"
        ;;
    *)
        echo "无效选择"
        ;;
esac

echo
echo -e "${GREEN}测试完成！${NC}"
echo

# 如果所有基础测试通过，显示成功信息
echo -e "${YELLOW}PVE Web界面验证:${NC}"
echo "1. 登录PVE Web界面"
echo "2. 选择VM ${VM_ID}"
echo "3. 查看 '选项' 标签页"
echo "4. 确认 'QEMU Guest Agent' 显示为 '是'"
echo "5. 测试 '关机' 按钮进行优雅关机"

echo
echo -e "${GREEN}🎉 macOS Guest Agent测试完成！${NC}" 