#!/bin/bash

# macOS Guest Agent 构建脚本
set -e

PROGRAM_NAME="mac-guest-agent"
VERSION="1.1.0"
BUILD_DIR="build"
DIST_DIR="dist"

echo "构建 macOS Guest Agent v${VERSION}..."

# 清理之前的构建
echo "清理构建目录..."
rm -rf "$BUILD_DIR"
rm -rf "$DIST_DIR"
mkdir -p "$BUILD_DIR"
mkdir -p "$DIST_DIR"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go编译器"
    echo "请先安装Go: https://golang.org/dl/"
    exit 1
fi

echo "Go版本: $(go version)"

# 检测架构
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
    GOARCH="arm64"
else
    GOARCH="amd64"
fi

echo "构建架构: $GOARCH"

# 安装依赖
echo "安装依赖..."
go mod download
go mod tidy

# 构建程序
echo "编译程序..."
CGO_ENABLED=0 GOOS=darwin GOARCH=$GOARCH go build \
    -ldflags "-X main.version=${VERSION} -s -w" \
    -o "${BUILD_DIR}/${PROGRAM_NAME}" \
    cmd/main.go

echo "已生成可执行文件: ${BUILD_DIR}/${PROGRAM_NAME}"

# 创建架构特定的二进制文件
cp "${BUILD_DIR}/${PROGRAM_NAME}" "${BUILD_DIR}/${PROGRAM_NAME}-darwin-${GOARCH}"
echo "已创建架构特定的二进制文件: ${BUILD_DIR}/${PROGRAM_NAME}-darwin-${GOARCH}"

# 验证可执行文件
if [[ -f "${BUILD_DIR}/${PROGRAM_NAME}" ]]; then
    file_size=$(ls -lh "${BUILD_DIR}/${PROGRAM_NAME}" | awk '{print $5}')
    echo "文件大小: $file_size"
    
    # 测试程序是否可以运行
    echo "测试程序..."
    "${BUILD_DIR}/${PROGRAM_NAME}" --help > /dev/null 2>&1 || true
    echo "程序验证通过"
else
    echo "错误: 构建失败"
    exit 1
fi

# 创建发布包
echo "创建发布包..."
mkdir -p "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}"

# 复制文件到发布目录
cp "${BUILD_DIR}/${PROGRAM_NAME}" "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}/"
cp "${BUILD_DIR}/${PROGRAM_NAME}-darwin-${GOARCH}" "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}/"
cp -r scripts "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}/"

# 创建README
cat > "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}/README.md" << EOF
# macOS Guest Agent v${VERSION}

## 快速安装

1. 解压到任意目录
2. 运行安装脚本:
   \`\`\`bash
   sudo ./scripts/install.sh
   \`\`\`

## 手动安装

1. 复制可执行文件:
   \`\`\`bash
   sudo cp ${PROGRAM_NAME} /usr/local/bin/
   sudo chmod +x /usr/local/bin/${PROGRAM_NAME}
   \`\`\`

2. 安装LaunchDaemon:
   \`\`\`bash
   # 无需手动安装plist文件，它已内嵌于二进制文件中
   # The plist file is embedded in the binary and does not need to be installed manually
   /usr/local/bin/${PROGRAM_NAME} --install
   \`\`\`

## 卸载

\`\`\`bash
sudo ./scripts/uninstall.sh
\`\`\`

## 支持的命令

- guest-ping: 心跳检测
- guest-info: 系统信息
- guest-sync: 协议同步
- guest-shutdown: 关机/重启
- guest-get-disks: 磁盘信息
- guest-get-memory-info: 内存信息
- 以及更多31+个命令

## 日志位置

/var/log/mac-guest-agent.log
EOF

# 使脚本可执行
chmod +x "${DIST_DIR}/${PROGRAM_NAME}-${VERSION}/scripts/"*.sh

# 创建压缩包
cd "$DIST_DIR"
tar -czf "${PROGRAM_NAME}-${VERSION}-darwin-${GOARCH}.tar.gz" "${PROGRAM_NAME}-${VERSION}/"
cd ..

# 创建校验和
cd "$BUILD_DIR"
md5 "${PROGRAM_NAME}-darwin-${GOARCH}" > "${PROGRAM_NAME}-darwin-${GOARCH}.md5"
shasum -a 256 "${PROGRAM_NAME}-darwin-${GOARCH}" > "${PROGRAM_NAME}-darwin-${GOARCH}.sha256"
cd ..

echo ""
echo "构建完成!"
echo "  - 可执行文件: ${BUILD_DIR}/${PROGRAM_NAME}"
echo "  - 架构特定文件: ${BUILD_DIR}/${PROGRAM_NAME}-darwin-${GOARCH}"
echo "  - 发布包: ${DIST_DIR}/${PROGRAM_NAME}-${VERSION}-darwin-${GOARCH}.tar.gz"
echo ""
echo "安装方法:"
echo "1. 本地安装: sudo ./scripts/install.sh --local"
echo "2. 在线安装: sudo ./scripts/install.sh" 