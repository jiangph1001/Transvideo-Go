# 视频转码管理系统 - Makefile 使用指南

## 快速开始（3个命令）

```bash
# 1. 检查环境 ⚠️
$ tools/check-docker.sh

# 2. 生成测试视频（自动生成 10 秒测试视频）
$ make demo-video

# 3. 启动服务 🚀
$ make run

# 服务启动后，打开浏览器访问 http://localhost:8080
```

## Makefile 命令分类

### 🎯 开发常用命令

| 命令 | 说明 | 使用场景 |
|------|------|----------|
| `make build` | 编译 Go 程序 | 修改代码后需要重新编译 |
| `make run` | 本地运行（自动创建测试目录） | 快速测试功能 |
| `make clean` | 清理编译产物 | 释放空间 |
| `make test` | 运行测试 | 代码开发 |

### 🐳 Docker 相关

| 命令 | 说明 | 使用场景 |
|------|------|----------|
| `make docker-build` | 构建 Docker 镜像 | 准备部署 |
| `make docker-run` | 运行 Docker 容器 | 生产环境 |
| `make docker-logs` | 查看容器日志 | 排查问题 |
| `make docker-stop` | 停止容器 | 停止服务 |

### 🎬 测试和演示

| 命令 | 说明 | 使用场景 |
|------|------|----------|
| `make demo-setup` | 创建测试目录 | 准备测试环境 |
| `make demo-video` | 生成测试视频（10秒） | 没有视频文件时 |
| `make demo` | 完整演示流程 | 第一次使用 |
| `tools/quick-test.sh video.mp4` | 快速测试 | 验证功能 |

### 🔧 工具

| 命令 | 说明 | 使用场景 |
|------|------|----------|
| `make help` | 显示所有命令 | 不记得命令时 |
| `tools/check-docker.sh` | 检查 Docker 环境 | 排查环境错误 |
| `make fmt` | 格式化 Go 代码 | 代码规范 |
| `make lint` | 代码检查 | 保证代码质量 |

---

## 典型工作流程

### 场景 1：本地开发和测试

```bash
# 1. 生成测试视频（只需要一次）
$ make demo-video

# 2. 修改代码...

# 3. 重新编译并运行
$ make run

# 4. 浏览器访问 http://localhost:8080

# 5. 输入文件名: test-video.mp4

# 6. 查看输出
$ ls -lh test-output/

# 7. 清理（可选）
$ make clean
```

### 场景 2：生产环境部署

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 文件（配置 INPUT_DIR 和 OUTPUT_DIR）
nano .env

# 2. 启动 Docker 容器
$ make docker-run

# 3. 查看日志
$ make docker-logs

# 4. 停止服务
$ make docker-stop
```

### 场景 3：测试现有视频

```bash
# 方法 1: 使用快速测试脚本
tools/quick-test.sh /path/to/your/video.mp4

# 方法 2: 手动步骤
cp /path/to/your/video.mp4 test-videos/
make run
# 然后在 Web 界面输入文件名
```

---

## 快捷命令

### 一键演示（适合第一次使用）

```bash
# 自动完成: 环境检查 + 生成视频 + 启动服务
$ make demo
```

### 不使用 Makefile 的纯脚本方式

如果你不想使用 make，可以直接运行：

```bash
# 给予执行权限
chmod +x run-local.sh

# 运行（自动完成编译、生成测试视频、启动服务）
./run-local.sh
```

---

## 环境要求

### 必须安装

- **Docker** - 用于运行转码容器
  - 安装: https://docs.docker.com/get-docker/

- **Go 1.21+** - 用于编译（本地运行时需要）
  - 安装: https://go.dev/doc/install

- **docker-compose** - 用于 Docker 部署
  - 通常随 Docker Desktop 安装

### 可选安装

- **ffmpeg** - 用于生成测试视频（`make demo-video`）
  - Ubuntu/Debian: `sudo apt-get install ffmpeg`
  - macOS: `brew install ffmpeg`
  - Windows: 下载安装包

- **golangci-lint** - 代码检查（`make lint`）
  - 安装: https://golangci-lint.run/usage/install/

---

## 常见问题

### Q: 运行 `make run` 时提示权限错误？

**A: 确保 Docker socket 有权限**

```bash
# 临时解决
sudo chmod 666 /var/run/docker.sock

# 永久解决（推荐）
sudo usermod -aG docker $USER
# 然后重新登录（或重启电脑）
```

### Q: `make demo-video` 提示找不到 ffmpeg？

**A: 安装 ffmpeg**

```bash
# Ubuntu/Debian
sudo apt-get install ffmpeg

# macOS
brew install ffmpeg
```

或者手动复制视频文件到 `test-videos/` 目录。

### Q: 服务启动后无法访问？

**A: 检查端口占用**

```bash
# 检查 8080 端口
netstat -tuln | grep 8080

# 或使用其他端口（修改代码中的端口号）
```

### Q: 如何调整并发数？

**A: 修改环境变量**

```bash
# 本地运行
MAX_CONCURRENCY=4 make run

# Docker 运行（修改 .env 文件）
# MAX_CONCURRENCY=4
```

### Q: 转码太慢？

**A: 调整参数**

```bash
# 降低质量（更快，文件更大）
# 在 Web 界面码率输入: 2M

# 提高预设（更快，质量稍低）
# 修改 service/docker_executor.go 中的 -preset 值
# 预设范围: 0-9，数字越小速度越慢质量越高
```

### Q: 输出文件太大？

**A: 降低码率**

```bash
# Web 界面码率输入:
# 500k = 500 kbps（文件小，质量较低）
# 1M   = 1 Mbps（平衡）
```

---

## 最佳实践

### ✅ 推荐做法

1. **本地开发**：使用 `make run` 快速测试
2. **配置文件**：复制 `.env.example` 到 `.env`，不要提交 `.env`
3. **测试视频**：使用 `make demo-video` 生成小视频（10秒）
4. **大文件处理**：使用 `MAX_CONCURRENCY=1`，避免 I/O 阻塞
5. **日志查看**：使用 `make docker-logs` 或 `docker-compose logs`

### ❌ 不推荐做法

- 不要在代码中硬编码路径
- 不要使用 root 用户运行（除非必要）
- 不要一次添加太多任务（系统资源限制）

---

## 典型场景示例

### 场景：批量转码多个视频

```bash
# 1. 将所有视频复制到测试目录
cp /path/to/videos/*.mp4 test-videos/

# 2. 启动服务
make run

# 3. 在 Web 界面依次添加任务
# 例如: video1.mp4, video2.mp4, video3.mp4

# 4. 查看进度
# Web 界面自动刷新（每5秒）
```

### 场景：调整转码质量

```bash
# 低质量（体积小，转码快）
# 码率: 500k

# 中等质量（平衡）
# 码率: 1M

# 高质量（体积大，转码慢）
# 码率: 2M
```

### 场景：使用 API 集成

```bash
#!/bin/bash

# 批量添加任务
for video in /path/to/videos/*.mp4; do
    filename=$(basename "$video")
    curl -X POST http://localhost:8080/api/tasks \
        -H "Content-Type: application/json" \
        -d "{\"input_file\": \"$filename\", \"bitrate\": \"1M\"}"
done
```

---

## 性能调优

### 根据 CPU 核心数设置并发

| CPU 核心数 | 建议 `MAX_CONCURRENCY` | 说明 |
|-----------|----------------------|------|
| 4 核      | 2                    | AV1 编码 CPU 占用高 |
| 8 核      | 3-4                  | 预留 50% CPU |
| 16 核     | 4-8                  | 根据实际负载调整 |
| 32 核     | 8-16                 | 可更高 |

### 监控系统资源

```bash
# 监控 CPU 和内存
htop

# 或者使用 docker stats（Docker 模式）
docker stats

# 查看日志
tail -f /path/to/log
```

---

## 故障排查

### 问题 1：任务一直 pending

**原因**：
- 所有 worker 都在忙
- 队列已满（1000 个任务）

**解决**：
```bash
# 1. 查看任务状态
make docker-logs

# 2. 增加并发数（修改 .env）
MAX_CONCURRENCY=4

# 3. 重启服务
make docker-stop && make docker-run
```

### 问题 2：转码失败（红色 failed）

**原因**：
- 输入文件不存在
- 文件格式不支持
- 磁盘空间不足
- 内存不足

**解决**：
```bash
# 1. 查看详细错误
make docker-logs

# 2. 检查输入文件
ls -lh test-videos/

# 3. 检查磁盘空间
df -h

# 4. 检查内存
free -h
```

### 问题 3：Docker 挂载失败

**原因**：
- Docker socket 权限不足
- 路径不存在

**解决**：
```bash
# 1. 检查 Docker socket
ls -l /var/run/docker.sock

# 2. 检查路径存在
ls -ld /path/to/input/dir
ls -ld /path/to/output/dir

# 3. 如果是远程 Docker，检查 DOCKER_HOST
env | grep DOCKER_HOST
```

---

## 相关文档

- [README.md](README.md) - 完整文档
- [QUICKSTART.md](QUICKSTART.md) - 快速开始指南
- [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - 项目结构说明
- [Makefile](Makefile) - 所有 make 命令源码

---

## 总结

### 最快上手（3步）

```bash
$ tools/check-docker.sh      # 检查环境
$ make demo-video            # 生成测试视频（如果有视频可跳过）
$ make run                   # 启动服务
```

### 常用命令记忆

```bash
# 开发
make build run clean

# Docker
make docker-*

# 测试
make demo*-video

# 工具
make help
```

### 关键路径

- 输入视频：`test-videos/`
- 输出视频：`test-output/`
- Web 界面：`http://localhost:8080`
- API：`http://localhost:8080/api/`
