-- 添加 attachments 字段到 messages 表
ALTER TABLE messages ADD COLUMN IF NOT EXISTS attachments JSONB DEFAULT '[]';

-- 更新现有的图片消息，从 voice_url 字段迁移到 attachments
UPDATE messages 
SET attachments = jsonb_build_array(
    jsonb_build_object(
        'id', gen_random_uuid()::text,
        'filename', content,
        'url', voice_url,
        'proxy_url', voice_url,
        'size', 0,
        'content_type', 'image/png'
    )
)
WHERE type = 'image' AND voice_url IS NOT NULL AND attachments = '[]';

-- 更新现有的文件消息，从 voice_url 字段迁移到 attachments
UPDATE messages 
SET attachments = jsonb_build_array(
    jsonb_build_object(
        'id', gen_random_uuid()::text,
        'filename', content,
        'url', voice_url,
        'proxy_url', voice_url,
        'size', 0,
        'content_type', 'application/octet-stream'
    )
)
WHERE type = 'file' AND voice_url IS NOT NULL AND attachments = '[]';
