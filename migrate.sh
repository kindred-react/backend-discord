#!/bin/bash

# 图片附件功能数据库迁移脚本
# 用途：为 messages 表添加 attachments 字段

set -e

echo "=========================================="
echo "Discord 后端 - 数据库迁移脚本"
echo "添加 attachments 字段到 messages 表"
echo "=========================================="
echo ""

# 读取 .env 文件中的数据库连接信息
if [ -f .env ]; then
    export $(cat .env | grep DATABASE_URL | xargs)
else
    echo "错误: 找不到 .env 文件"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo "错误: DATABASE_URL 未设置"
    exit 1
fi

echo "数据库连接: $DATABASE_URL"
echo ""

# 提取数据库连接参数
# 格式: postgres://user:password@host:port/dbname?sslmode=disable
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
DB_PASS=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')

echo "数据库信息:"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo "  数据库: $DB_NAME"
echo "  用户: $DB_USER"
echo ""

# 确认执行
read -p "是否继续执行迁移? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消迁移"
    exit 0
fi

echo ""
echo "开始执行迁移..."
echo ""

# 执行迁移 SQL
PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f add_attachments_column.sql

if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✅ 迁移成功完成！"
    echo "=========================================="
    echo ""
    echo "下一步:"
    echo "1. 重启后端服务"
    echo "2. 测试图片上传和显示功能"
    echo ""
else
    echo ""
    echo "=========================================="
    echo "❌ 迁移失败"
    echo "=========================================="
    echo ""
    echo "请检查错误信息并重试"
    exit 1
fi
