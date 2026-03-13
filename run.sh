#!/bin/bash

# Discord 后端启动脚本（优化版）

cd "$(dirname "$0")"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Discord 后端启动脚本${NC}"
echo ""

# 检查是否需要编译
NEED_BUILD=false

if [ ! -f bin/server ]; then
    echo -e "${YELLOW}📦 首次运行，需要编译...${NC}"
    NEED_BUILD=true
elif [ cmd/server/main.go -nt bin/server ]; then
    echo -e "${YELLOW}📦 检测到代码更新，重新编译...${NC}"
    NEED_BUILD=true
else
    # 检查其他 Go 文件是否有更新
    if [ -n "$(find internal -name '*.go' -newer bin/server 2>/dev/null)" ]; then
        echo -e "${YELLOW}📦 检测到代码更新，重新编译...${NC}"
        NEED_BUILD=true
    else
        echo -e "${GREEN}✅ 使用已编译的版本（跳过编译）${NC}"
    fi
fi

if [ "$NEED_BUILD" = true ]; then
    mkdir -p bin
    
    # 显示编译进度
    echo -e "${BLUE}⏳ 正在编译...${NC}"
    
    # 使用优化的编译选项
    # -ldflags="-s -w" 减小二进制文件大小
    # -trimpath 移除文件路径信息
    go build -ldflags="-s -w" -trimpath -o bin/server cmd/server/main.go
    
    if [ $? -ne 0 ]; then
        echo -e "${RED}❌ 编译失败${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 编译成功${NC}"
    
    # 显示二进制文件大小
    SIZE=$(du -h bin/server | cut -f1)
    echo -e "${BLUE}📊 二进制文件大小: ${SIZE}${NC}"
fi

echo ""
echo -e "${GREEN}🎯 启动服务器...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 运行服务器
./bin/server
