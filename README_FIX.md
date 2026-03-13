# 图片缩略图显示修复说明

## 问题描述
文本消息流中的图片无法显示缩略图，也无法获取完整图片。

## 问题原因
数据库 `messages` 表缺少 `attachments` 字段，导致图片附件信息无法持久化。虽然上传图片时返回了附件信息，但刷新消息列表后，从数据库查询的消息没有附件数据。

## 修复内容

### 1. 数据库修改
- 在 `messages` 表中添加 `attachments` JSONB 字段
- 更新数据库初始化脚本 (`internal/database/database.go`)

### 2. 后端服务修改
- 修改 `MessageService.CreateWithType()` 方法，添加 `CreateWithAttachments()` 方法
- 修改 `MessageService.GetByID()` 和 `GetByChannelID()` 方法，支持读取 attachments
- 修改 `MessageService.CreateCallRecord()` 和 `CreateVoiceMessage()` 方法
- 修改 `FileHandler.UploadImage()` 和 `UploadFile()` 方法，使用新的 `CreateWithAttachments()` 方法

### 3. 前端已支持
前端代码已经支持 `attachments` 字段，无需修改。

## 部署步骤

### 步骤 1: 备份数据库
```bash
# 如果使用 PostgreSQL
pg_dump -U your_user -d your_database > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 步骤 2: 运行数据库迁移
```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 方法 1: 使用 psql 命令
psql -U your_user -d your_database -f add_attachments_column.sql

# 方法 2: 或者在 PostgreSQL 客户端中执行 add_attachments_column.sql 文件内容
```

### 步骤 3: 重启后端服务
```bash
# 停止当前运行的服务
# 然后重新启动
cd /Users/apple/Documents/trae_projects/backend-discord
go run cmd/server/main.go
```

### 步骤 4: 测试
1. 登录应用
2. 选择一个频道
3. 上传一张图片
4. 验证图片缩略图是否正常显示
5. 点击图片，验证是否能查看完整图片
6. 刷新页面，验证图片是否仍然显示

## 验证清单
- [ ] 数据库成功添加 `attachments` 字段
- [ ] 后端服务成功启动，无错误日志
- [ ] 上传图片后，消息列表显示缩略图
- [ ] 点击缩略图可以查看完整图片
- [ ] 刷新页面后，图片仍然正常显示
- [ ] 上传文件后，显示文件信息和下载按钮

## 技术细节

### 数据库字段
```sql
attachments JSONB DEFAULT '[]'
```

### Attachments 数据结构
```json
[
  {
    "id": "uuid",
    "filename": "image.png",
    "url": "/uploads/images/uuid_timestamp.png",
    "proxy_url": "/uploads/images/uuid_timestamp.png",
    "size": 12345,
    "width": 800,
    "height": 600,
    "content_type": "image/png"
  }
]
```

### API 响应示例
```json
{
  "id": "message-uuid",
  "content": "image.png",
  "type": "image",
  "attachments": [
    {
      "id": "attachment-uuid",
      "filename": "image.png",
      "url": "/uploads/images/uuid_timestamp.png",
      "size": 12345,
      "width": 800,
      "height": 600,
      "content_type": "image/png"
    }
  ],
  "author": {...},
  "created_at": "2024-01-01T00:00:00Z"
}
```

## 回滚方案
如果出现问题，可以回滚：

```sql
-- 删除 attachments 字段
ALTER TABLE messages DROP COLUMN IF EXISTS attachments;

-- 恢复备份
psql -U your_user -d your_database < backup_YYYYMMDD_HHMMSS.sql
```

## 注意事项
1. 确保 `/uploads/images/` 和 `/uploads/files/` 目录存在且有写权限
2. 确保静态文件服务配置正确 (`r.Static("/uploads", "./uploads")`)
3. 如果使用反向代理（如 Nginx），确保正确配置了静态文件路径
