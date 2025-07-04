#!/bin/bash
# 警告：这是一个具有破坏性的操作，它将重写整个 'main' 分支的历史记录。
# 在运行此脚本之前，请确保您了解其后果。
#
# 该脚本会将所有提交压缩成一个单一的初始提交。

set -e

# 1. 创建一个没有历史记录的孤立分支
echo "步骤 1/5: 创建孤立分支 'temp_branch'..."
git checkout --orphan temp_branch

# 2. 添加所有文件到暂存区
echo "步骤 2/5: 添加所有文件..."
git add -A

# 3. 创建一个新的、单一的提交
echo "步骤 3/5: 创建单一的初始提交..."
git commit -m "feat: compact initial commit"

# 4. 删除旧的main分支
echo "步骤 4/5: 删除旧的 'main' 分支..."
git branch -D main

# 5. 将当前分支重命名为main
echo "步骤 5/5: 将当前分支重命名为 'main'..."
git branch -m main
git push -f origin main