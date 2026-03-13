#!/bin/bash

# 测试脚本 - 验证 attachments 字段是否正确添加

set -e

echo "=========================================="
echo "验证数据库迁移"
echo "=========================================="
echo ""

# 读取 .env 文件
if [ -f .env ]; then
    export $(cat .env | grep DATABASE_URL | xargs)
else
    echo "错误: 找不到 .env 文件"
    exit 1
fi

# 提取数据库连接参数
DB_USER=$(echo $DATABASE_URL | sed -n 's/.*:\/\/\([^:]*\):.*/\1/p')
DB_PASS=$(echo $DATABASE_URL | sed -n 's/.*:\/\/[^:]*:\([^@]*\)@.*/\1/p')
DB_HOST=$(echo $DATABASE_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
DB_PORT=$(echo $DATABASE_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')

echo "检查 messages 表结构..."
echo ""

# 检查 attachments 字段是否存在
RESULT=$(PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT column_name, data_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'messages' 
AND column_name = 'attachments';
")

if [ -z "$RESULT" ]; then
    echo "❌ attachments 字段不存在"
    echo ""
    echo "请先运行迁移脚本:"
    echo "  ./migrate.sh"
    exit 1
else
    echo "✅ attachments 字段已存在"
    echo "$RESULT"
    echo ""
fi

# 检查是否有图片消息
echo "检查图片消息..."
COUNT=$(PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT COUNT(*) FROM messages WHERE type = 'image';
")

echo "图片消息数量: $COUNT"
echo ""

if [ "$COUNT" -gt 0 ]; then
    echo "查看最近的图片消息 (前3条):"
    PGPASSWORD=$DB_PASS psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        id, 
        content, 
        type, 
        jsonb_array_length(attachments) as attachment_count,
        created_at 
    FROM messages 
    WHERE type = 'image' 
    ORDER BY created_at DESC 
    LIMIT 3;
    "
    echo ""
fi

echo "=========================================="
echo "✅ 验证完成"
echo "=========================================="
echo ""
echo "如果看到 attachments 字段存在，说明迁移成功！"
echo ""
