#!/bin/bash

# Discord 后端开发模式（自动重新编译）

cd "$(dirname "$0")"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🔧 Discord 后端开发模式${NC}"
echo -e "${YELLOW}提示: 修改代码后，按 Ctrl+C 停止，然后重新运行此脚本${NC}"
echo ""

# 强制重新编译
echo -e "${BLUE}🔨 编译中...${NC}"
mkdir -p bin

# 编译并显示时间
START_TIME=$(date +%s)
go build -o bin/server cmd/server/main.go

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 编译失败${NC}"
    exit 1
fi

END_TIME=$(date +%s)
COMPILE_TIME=$((END_TIME - START_TIME))

echo -e "${GREEN}✅ 编译完成 (${COMPILE_TIME}秒)${NC}"
echo ""
echo -e "${GREEN}🚀 启动服务器...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 运行服务器
./bin/server
