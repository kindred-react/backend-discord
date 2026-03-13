# 图片缩略图显示修复 - 完整部署指南

## 📋 问题总结

**症状**: 
- 上传图片后，文本消息流中无法显示图片缩略图
- 无法点击查看完整图片
- 刷新页面后，图片信息丢失

**根本原因**:
数据库 `messages` 表缺少 `attachments` 字段，导致图片附件信息无法持久化存储。

## 🔧 修复内容

### 已修改的文件

1. **backend-discord/internal/database/database.go**
   - 在 `messages` 表定义中添加 `attachments JSONB DEFAULT '[]'` 字段

2. **backend-discord/internal/services/services.go**
   - 新增 `CreateWithAttachments()` 方法
   - 修改 `CreateWithType()` 调用新方法
   - 修改 `GetByID()` 读取 attachments
   - 修改 `GetByChannelID()` 读取 attachments
   - 修改 `CreateCallRecord()` 和 `CreateVoiceMessage()` 支持 attachments

3. **backend-discord/internal/handlers/file.go**
   - 修改 `UploadImage()` 使用 `CreateWithAttachments()`
   - 修改 `UploadFile()` 使用 `CreateWithAttachments()`

### 新增的文件

1. **add_attachments_column.sql** - 数据库迁移脚本
2. **migrate.sh** - 自动化迁移脚本
3. **verify_migration.sh** - 验证脚本
4. **README_FIX.md** - 详细修复说明
5. **DEPLOYMENT_GUIDE.md** - 本文档

## 🚀 部署步骤

### 前置条件

- PostgreSQL 已安装并运行
- 有数据库访问权限
- Go 环境已配置
- 后端服务当前未运行（或准备重启）

### 步骤 1: 备份数据库（重要！）

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 创建备份目录
mkdir -p backups

# 备份数据库
pg_dump -U postgres -d discord > backups/backup_$(date +%Y%m%d_%H%M%S).sql

# 验证备份文件
ls -lh backups/
```

### 步骤 2: 执行数据库迁移

**方法 A: 使用自动化脚本（推荐）**

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 执行迁移
./migrate.sh

# 按提示输入 'y' 确认
```

**方法 B: 手动执行 SQL**

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 使用 psql 执行
psql -U postgres -d discord -f add_attachments_column.sql
```

### 步骤 3: 验证迁移

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 运行验证脚本
./verify_migration.sh
```

**预期输出**:
```
✅ attachments 字段已存在
 attachments | jsonb | '[]'::jsonb
```

### 步骤 4: 重启后端服务

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 如果服务正在运行，先停止（Ctrl+C）

# 重新编译并运行
go run cmd/server/main.go
```

**预期输出**:
```
Database connected successfully
Database tables created successfully
Server starting on :8080
```

### 步骤 5: 测试功能

#### 5.1 测试图片上传

1. 打开前端应用 (http://localhost:5173 或相应端口)
2. 登录账号
3. 选择一个文字频道
4. 点击 "+" 按钮上传图片
5. 选择一张图片（PNG/JPG/GIF，小于 5MB）

**预期结果**:
- ✅ 图片上传成功
- ✅ 消息列表显示图片缩略图
- ✅ 缩略图保持原始比例（最大 400x300px）

#### 5.2 测试图片查看

1. 点击缩略图
2. 应该打开全屏模态框显示完整图片

**预期结果**:
- ✅ 显示完整尺寸图片
- ✅ 显示图片尺寸信息
- ✅ 可以下载原图
- ✅ 按 ESC 或点击背景关闭

#### 5.3 测试持久化

1. 刷新浏览器页面（F5）
2. 重新进入相同频道

**预期结果**:
- ✅ 之前上传的图片仍然显示
- ✅ 缩略图正常加载
- ✅ 可以点击查看完整图片

#### 5.4 测试文件上传

1. 点击 "+" 按钮上传文件
2. 选择一个非图片文件（PDF/DOC/TXT 等）

**预期结果**:
- ✅ 文件上传成功
- ✅ 显示文件图标和信息
- ✅ 显示文件大小
- ✅ 可以点击下载

## 🔍 故障排查

### 问题 1: 迁移脚本执行失败

**错误**: `psql: command not found`

**解决方案**:
```bash
# macOS
brew install postgresql

# 或者使用完整路径
/Applications/Postgres.app/Contents/Versions/latest/bin/psql ...
```

### 问题 2: 数据库连接失败

**错误**: `connection refused`

**解决方案**:
```bash
# 检查 PostgreSQL 是否运行
pg_isready

# 启动 PostgreSQL (macOS)
brew services start postgresql

# 或者
pg_ctl -D /usr/local/var/postgres start
```

### 问题 3: 权限错误

**错误**: `permission denied for table messages`

**解决方案**:
```sql
-- 以超级用户身份连接
psql -U postgres -d discord

-- 授予权限
GRANT ALL PRIVILEGES ON TABLE messages TO your_user;
```

### 问题 4: 图片仍然不显示

**检查清单**:

1. **验证数据库字段**:
```bash
./verify_migration.sh
```

2. **检查后端日志**:
```bash
# 查看是否有错误信息
# 应该看到成功的上传日志
```

3. **检查网络请求**:
- 打开浏览器开发者工具 (F12)
- 切换到 Network 标签
- 上传图片
- 检查 `/api/files/image` 请求
- 查看响应中是否包含 `attachments` 字段

4. **检查静态文件服务**:
```bash
# 确认文件已上传
ls -la uploads/images/

# 测试直接访问
curl http://localhost:8080/uploads/images/文件名.png
```

### 问题 5: 旧消息没有附件信息

这是正常的。迁移脚本会尝试从 `voice_url` 字段迁移数据，但如果旧消息的 `voice_url` 为空，则无法恢复。

**解决方案**: 重新上传图片即可。

## 📊 数据库结构

### messages 表结构（更新后）

```sql
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    voice_url VARCHAR(255),
    duration INTEGER,
    attachments JSONB DEFAULT '[]',  -- 新增字段
    embeds JSONB DEFAULT '[]',
    reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### attachments 字段示例

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "screenshot.png",
    "url": "/uploads/images/550e8400-e29b-41d4-a716-446655440000_1710345678.png",
    "proxy_url": "/uploads/images/550e8400-e29b-41d4-a716-446655440000_1710345678.png",
    "size": 245678,
    "width": 1920,
    "height": 1080,
    "content_type": "image/png"
  }
]
```

## 🔄 回滚方案

如果迁移后出现严重问题，可以回滚：

```bash
cd /Users/apple/Documents/trae_projects/backend-discord

# 停止后端服务

# 恢复数据库备份
psql -U postgres -d discord < backups/backup_YYYYMMDD_HHMMSS.sql

# 或者只删除 attachments 字段
psql -U postgres -d discord -c "ALTER TABLE messages DROP COLUMN IF EXISTS attachments;"

# 恢复代码（使用 git）
git checkout HEAD -- internal/database/database.go
git checkout HEAD -- internal/services/services.go
git checkout HEAD -- internal/handlers/file.go

# 重新编译运行
go run cmd/server/main.go
```

## ✅ 验证清单

部署完成后，请确认以下所有项目：

- [ ] 数据库迁移成功执行
- [ ] `attachments` 字段已添加到 `messages` 表
- [ ] 后端服务成功启动，无错误日志
- [ ] 可以上传图片
- [ ] 图片缩略图正常显示
- [ ] 点击缩略图可以查看完整图片
- [ ] 刷新页面后图片仍然显示
- [ ] 可以上传文件
- [ ] 文件信息正常显示
- [ ] 可以下载文件
- [ ] WebSocket 消息推送正常工作

## 📞 支持

如果遇到问题：

1. 查看本文档的"故障排查"部分
2. 检查后端日志输出
3. 检查浏览器控制台错误
4. 验证数据库连接和表结构

## 📝 更新日志

**2024-03-13**
- 添加 `attachments` 字段到 `messages` 表
- 更新消息创建和查询逻辑
- 修复图片缩略图显示问题
- 修复图片持久化问题
