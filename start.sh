#!/bin/bash

# Discord 后端快速启动脚本（极速版）
# 此脚本会预先下载依赖，确保最快启动速度

cd "$(dirname "$0")"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}⚡ Discord 后端极速启动${NC}"
echo ""

# 步骤 1: 确保依赖已下载
if [ ! -d "$HOME/go/pkg/mod/github.com/gin-gonic" ]; then
    echo -e "${YELLOW}📥 首次运行，下载依赖...${NC}"
    go mod download
    echo -e "${GREEN}✅ 依赖下载完成${NC}"
fi

# 步骤 2: 检查是否需要编译
if [ ! -f bin/server ]; then
    echo -e "${YELLOW}🔨 首次编译...${NC}"
    mkdir -p bin
    
    # 使用最快的编译选项
    go build -ldflags="-s -w" -trimpath -o bin/server cmd/server/main.go
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ 编译失败${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 编译完成${NC}"
else
    echo -e "${GREEN}✅ 使用缓存的二进制文件${NC}"
fi

echo ""
echo -e "${GREEN}🚀 启动服务器（端口 8080）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 运行服务器
exec ./bin/server
