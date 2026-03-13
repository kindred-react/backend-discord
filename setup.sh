#!/bin/bash

# 预下载所有依赖（只需运行一次）

cd "$(dirname "$0")"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}📦 预下载 Go 依赖${NC}"
echo -e "${YELLOW}这个脚本只需要运行一次${NC}"
echo ""

echo -e "${BLUE}⏳ 下载依赖中...${NC}"
go mod download

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ 所有依赖已下载到缓存${NC}"
    echo ""
    echo -e "${BLUE}缓存位置:${NC}"
    go env GOMODCACHE
    echo ""
    echo -e "${GREEN}现在可以使用 ./start.sh 快速启动了！${NC}"
else
    echo -e "${RED}❌ 下载失败${NC}"
    exit 1
fi
