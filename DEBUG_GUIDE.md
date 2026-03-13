# 🔍 图片显示问题调试指南

## 当前状态

✅ 数据库迁移成功 - `attachments` 字段已添加  
✅ 后端代码已更新  
✅ 后端服务正在运行  
✅ 图片文件存在于 `uploads/images/`  
✅ 静态文件可以直接访问  
❓ 前端显示图片 - 需要验证

## 快速诊断步骤

### 步骤 1: 检查浏览器控制台

1. 打开前端应用 (http://localhost:5173)
2. 按 F12 打开开发者工具
3. 切换到 Console 标签
4. 查看是否有错误信息

**常见错误**:
- `404 Not Found` - 图片路径错误
- `CORS error` - 跨域问题
- `Failed to load resource` - 资源加载失败

### 步骤 2: 检查网络请求

1. 在开发者工具中切换到 Network 标签
2. 刷新页面
3. 查找 `/api/channels/.../messages` 请求
4. 点击该请求，查看 Response

**检查点**:
- Response 中是否包含 `attachments` 字段？
- `attachments[0].url` 的值是什么？
- 是否有图片请求（如 `/uploads/images/...`）？
- 图片请求的状态码是什么？

### 步骤 3: 使用测试页面

1. 打开浏览器访问: `http://localhost:8080/test_images.html`
2. 点击 "点击获取消息" 按钮
3. 查看返回的数据

**预期结果**:
- 应该看到消息列表
- 图片类型的消息应该有 `attachments` 数据
- 应该能看到图片缩略图

### 步骤 4: 直接测试图片访问

在浏览器中直接访问:
```
http://localhost:8080/uploads/images/b665ac4c-3619-46e6-80ed-7785fb61949e_1773415819.png
```

**预期结果**: 应该能看到图片

### 步骤 5: 检查前端代码

打开浏览器开发者工具的 Console，运行:

```javascript
// 检查 localStorage 中的 token
console.log('Token:', localStorage.getItem('token'));

// 手动获取消息
fetch('http://localhost:8080/api/channels/aaaaaaaa-aaaa-aaaa-aaaa-111111111111/messages?limit=1&offset=0', {
  headers: {
    'Authorization': 'Bearer ' + localStorage.getItem('token')
  }
})
.then(r => r.json())
.then(data => {
  console.log('Messages:', data);
  const msg = data.messages[0];
  console.log('First message type:', msg.type);
  console.log('First message attachments:', msg.attachments);
});
```

## 可能的问题和解决方案

### 问题 1: 前端缓存了旧数据

**症状**: 刷新页面后仍然看不到图片

**解决方案**:
```bash
# 清除浏览器缓存
# Chrome: Ctrl+Shift+Delete (Windows) 或 Cmd+Shift+Delete (Mac)
# 或者使用无痕模式: Ctrl+Shift+N (Windows) 或 Cmd+Shift+N (Mac)
```

### 问题 2: 前端代码未更新

**症状**: 前端代码中没有处理 attachments

**解决方案**:
```bash
cd /Users/apple/Documents/trae_projects/react-discord

# 重启前端开发服务器
# 按 Ctrl+C 停止
# 然后重新运行
npm run dev
```

### 问题 3: API 返回的数据没有 attachments

**症状**: Network 标签中看到的响应没有 attachments 字段

**解决方案**:
```bash
# 重启后端服务
cd /Users/apple/Documents/trae_projects/backend-discord

# 停止当前服务 (Ctrl+C)
# 清除 Go 缓存
go clean -cache

# 重新运行
go run cmd/server/main.go
```

### 问题 4: 图片 URL 路径错误

**症状**: 图片请求返回 404

**检查**:
- attachments[0].url 的值是否以 `/uploads/images/` 开头？
- 文件名是否正确？

**解决方案**:
```bash
# 检查实际的图片文件
ls -la /Users/apple/Documents/trae_projects/backend-discord/uploads/images/

# 检查数据库中的 URL
cd /Users/apple/Documents/trae_projects/backend-discord
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d discord -c "SELECT id, type, attachments->0->>'url' as image_url FROM messages WHERE type = 'image' LIMIT 3;"
```

### 问题 5: MessageItem 组件没有正确渲染

**症状**: 控制台没有错误，但图片不显示

**检查**:
1. 打开 React DevTools
2. 找到 MessageItem 组件
3. 查看 props 中的 message.attachments

**解决方案**: 检查 MessageItem.tsx 中的条件判断

## 完整的重启流程

如果以上都不行，尝试完全重启：

```bash
# 1. 停止所有服务
# 前端: Ctrl+C
# 后端: Ctrl+C

# 2. 清除缓存
cd /Users/apple/Documents/trae_projects/backend-discord
go clean -cache

cd /Users/apple/Documents/trae_projects/react-discord
rm -rf node_modules/.vite

# 3. 重启后端
cd /Users/apple/Documents/trae_projects/backend-discord
go run cmd/server/main.go

# 4. 重启前端（新终端）
cd /Users/apple/Documents/trae_projects/react-discord
npm run dev

# 5. 清除浏览器缓存并刷新
```

## 验证清单

完成以下检查：

- [ ] 浏览器控制台没有错误
- [ ] Network 标签中 messages API 返回了 attachments 数据
- [ ] 图片文件可以直接访问 (http://localhost:8080/uploads/images/...)
- [ ] test_images.html 页面可以显示图片
- [ ] React DevTools 中 MessageItem 的 props 包含 attachments
- [ ] 前端页面显示图片缩略图

## 获取帮助

如果问题仍然存在，请提供以下信息：

1. 浏览器控制台的错误截图
2. Network 标签中 messages API 的响应数据
3. 后端日志的最后 50 行
4. test_images.html 页面的显示结果

---

**提示**: 最常见的问题是浏览器缓存或服务未重启。建议先尝试完全重启流程。
