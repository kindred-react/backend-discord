#!/bin/bash

# Discord 后端快速启动脚本
cd "$(dirname "$0")"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}⚡ Discord 后端启动${NC}"
echo ""

# 步骤 1: 确保依赖已下载
if [ ! -d "$HOME/go/pkg/mod/github.com/gin-gonic" ]; then
    echo -e "${YELLOW}📥 首次运行，下载依赖...${NC}"
    go mod download
    echo -e "${GREEN}✅ 依赖下载完成${NC}"
fi

# 步骤 2: 每次重新编译，确保代码是最新的
echo -e "${YELLOW}🔨 编译中...${NC}"
mkdir -p bin

# 使用临时 GOCACHE 绕过缓存权限问题
export GOCACHE=$(mktemp -d)
go build -ldflags="-s -w" -trimpath -o bin/server cmd/server/main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 编译完成${NC}"
echo ""
echo -e "${GREEN}🚀 启动服务器（端口 8080）${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 运行服务器
exec ./bin/server
