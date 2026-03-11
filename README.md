# Discord Backend Clone

基于 Go + Gin + PostgreSQL + WebSocket 复刻 Discord 后端核心功能。

## 技术栈

- **语言**: Go 1.21
- **Web框架**: Gin
- **数据库**: PostgreSQL
- **WebSocket**: Gorilla WebSocket
- **认证**: JWT
- **密码加密**: bcrypt

## 项目结构

```
discord-backend/
├── cmd/server/          # 主入口
├── config/              # 配置
├── internal/
│   ├── database/        # 数据库连接和初始化
│   ├── handlers/        # HTTP 处理器
│   ├── middleware/      # 中间件 (JWT 认证)
│   ├── models/          # 数据模型
│   ├── services/        # 业务逻辑
│   └── websocket/       # WebSocket 处理
├── bin/                 # 编译输出
├── go.mod
└── README.md
```

## 功能特性

### REST API

- **用户认证**
  - `POST /api/auth/register` - 注册
  - `POST /api/auth/login` - 登录
  - `GET /api/auth/me` - 获取当前用户

- **服务器 (Guild)**
  - `POST /api/guilds` - 创建服务器
  - `GET /api/guilds` - 获取用户服务器列表
  - `GET /api/guilds/:id` - 获取服务器详情
  - `DELETE /api/guilds/:id` - 删除服务器
  - `GET /api/guilds/:id/members` - 获取成员列表

- **频道 (Channel)**
  - `POST /api/guilds/:guildId/channels` - 创建频道
  - `GET /api/guilds/:guildId/channels` - 获取服务器频道列表
  - `GET /api/channels/:id` - 获取频道详情
  - `DELETE /api/channels/:id` - 删除频道

- **消息 (Message)**
  - `GET /api/channels/:channelId/messages` - 获取消息历史
  - `POST /api/channels/:channelId/messages` - 发送消息
  - `DELETE /api/channels/messages/:id` - 删除消息

- **语音频道 (Voice)**
  - `POST /api/voice/join` - 加入语音频道
  - `POST /api/voice/leave` - 离开语音频道
  - `GET /api/voice/:channelId` - 获取频道在线用户

### WebSocket 实时通信

连接地址: `ws://localhost:8080/ws?user_id={user_id}`

**事件列表:**

- `message:send` - 发送消息
- `message:new` - 新消息推送
- `channel:voice:join` - 用户加入语音频道
- `channel:voice:leave` - 用户离开语音频道
- `join_channel` - 加入频道房间
- `leave_channel` - 离开频道房间
- `typing:start` - 开始输入

**WebSocket 消息格式:**

```json
{
  "event": "message:send",
  "data": {
    "channel_id": "xxx-xxx-xxx",
    "content": "Hello World!"
  }
}
```

## 快速开始

### 1. 环境要求

- Go 1.21+
- PostgreSQL 14+

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件:

```env
SERVER_PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/discord?sslmode=disable
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRE_HOURS=168
```

### 3. 创建数据库

```bash
createdb discord
```

### 4. 运行服务

```bash
# 开发模式
go run ./cmd/server

# 或者运行编译后的二进制
./bin/server
```

### 5. 测试 API

```bash
# 注册用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 使用 token 访问受保护路由
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## WebSocket 测试示例

使用 JavaScript 连接:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?user_id=USER_ID');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

// 发送消息
ws.send(JSON.stringify({
  event: 'message:send',
  data: {
    channel_id: 'CHANNEL_ID',
    content: 'Hello!'
  }
}));

// 加入频道房间
ws.send(JSON.stringify({
  event: 'join_channel',
  data: {
    channel_id: 'CHANNEL_ID'
  }
}));

// 加入语音频道
ws.send(JSON.stringify({
  event: 'channel:voice:join',
  data: {
    channel_id: 'VOICE_CHANNEL_ID',
    guild_id: 'GUILD_ID'
  }
}));
```

## 许可证

MIT
