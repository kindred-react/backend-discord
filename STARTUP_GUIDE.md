# 🚀 后端启动脚本使用指南

## 📋 可用脚本

### 1. `./setup.sh` - 首次设置（只运行一次）⭐

**用途**: 预下载所有 Go 依赖到缓存

**使用场景**: 
- 第一次克隆项目后
- 清理了 Go 缓存后

**命令**:
```bash
cd /Users/apple/Documents/trae_projects/backend-discord
./setup.sh
```

**效果**: 
- 下载所有依赖到 `~/go/pkg/mod`
- 之后启动不会再下载

---

### 2. `./start.sh` - 快速启动（推荐日常使用）⭐⭐⭐

**用途**: 最快速度启动服务器

**特点**:
- ✅ 使用已编译的二进制文件
- ✅ 不重新编译（除非是首次运行）
- ✅ 启动速度：< 1 秒

**使用场景**: 
- 代码没有改变时
- 只是重启服务器

**命令**:
```bash
cd /Users/apple/Documents/trae_projects/backend-discord
./start.sh
```

---

### 3. `./dev.sh` - 开发模式（代码改变后使用）⭐⭐

**用途**: 每次都重新编译

**特点**:
- ✅ 强制重新编译
- ✅ 显示编译时间
- ✅ 适合开发调试

**使用场景**: 
- 修改了代码
- 需要测试新功能

**命令**:
```bash
cd /Users/apple/Documents/trae_projects/backend-discord
./dev.sh
```

---

### 4. `./run.sh` - 智能启动（自动检测）⭐⭐

**用途**: 自动检测是否需要重新编译

**特点**:
- ✅ 检测代码是否改变
- ✅ 只在需要时编译
- ✅ 平衡速度和便利性

**使用场景**: 
- 不确定代码是否改变
- 想要自动化处理

**命令**:
```bash
cd /Users/apple/Documents/trae_projects/backend-discord
./run.sh
```

---

## 🎯 推荐工作流程

### 首次使用

```bash
# 1. 预下载依赖（只需一次）
./setup.sh

# 2. 首次编译并启动
./start.sh
```

### 日常开发

```bash
# 如果没有修改代码，只是重启
./start.sh

# 如果修改了代码
./dev.sh
```

---

## ⚡ 性能对比

| 脚本 | 首次启动 | 后续启动 | 是否重新编译 |
|------|---------|---------|-------------|
| `go run` | 10-15秒 | 10-15秒 | 每次都编译 |
| `./setup.sh` + `./start.sh` | 10秒 | <1秒 | 只编译一次 |
| `./dev.sh` | 5-10秒 | 5-10秒 | 每次都编译 |
| `./run.sh` | 5-10秒 | <1秒 | 智能检测 |

---

## 🔧 故障排查

### 问题 1: 启动很慢，显示 "downloading"

**原因**: Go 在验证依赖或真的在下载

**解决方案**:
```bash
# 运行一次 setup.sh
./setup.sh

# 之后使用 start.sh
./start.sh
```

### 问题 2: 修改代码后没有生效

**原因**: 使用了缓存的二进制文件

**解决方案**:
```bash
# 使用 dev.sh 强制重新编译
./dev.sh

# 或者删除旧的二进制文件
rm bin/server
./start.sh
```

### 问题 3: 权限错误

**原因**: 脚本没有执行权限

**解决方案**:
```bash
chmod +x setup.sh start.sh dev.sh run.sh
```

---

## 📊 启动时间优化

### 优化前（使用 `go run`）
```
启动时间: 10-15 秒
原因: 每次都编译 + 验证依赖
```

### 优化后（使用 `./start.sh`）
```
首次: 10 秒（编译）
后续: < 1 秒（直接运行）
提升: 10-15 倍
```

---

## 💡 高级技巧

### 技巧 1: 创建别名

在 `~/.zshrc` 中添加：
```bash
alias discord-start='cd /Users/apple/Documents/trae_projects/backend-discord && ./start.sh'
alias discord-dev='cd /Users/apple/Documents/trae_projects/backend-discord && ./dev.sh'
```

然后在任何目录都可以：
```bash
discord-start  # 快速启动
discord-dev    # 开发模式
```

### 技巧 2: 后台运行

```bash
# 后台运行
./start.sh &

# 查看日志
tail -f nohup.out

# 停止服务
pkill -f "bin/server"
```

### 技巧 3: 自动重启（开发时）

安装 `air`（Go 的热重载工具）：
```bash
go install github.com/cosmtrek/air@latest

# 使用 air 启动（代码改变自动重启）
air
```

---

## 🎉 总结

**最佳实践**:
1. 首次使用运行 `./setup.sh`
2. 日常开发使用 `./start.sh`（最快）
3. 修改代码后使用 `./dev.sh`
4. 不确定时使用 `./run.sh`（智能）

**启动速度**: 从 10-15 秒优化到 < 1 秒！🚀
