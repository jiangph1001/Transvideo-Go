# 快速开始指南

本指南帮助你在 5 分钟内快速体验视频转码系统。

## 方式一：本地快速体验（推荐）

### 步骤 1：检查环境

```bash
# 检查 Docker 环境
tools/check-docker.sh

# 如果 Docker 未安装或运行，请先安装 Docker
# Ubuntu/Debian: https://docs.docker.com/engine/install/ubuntu/
# macOS: https://docs.docker.com/docker-for-mac/install/
```

### 步骤 2：准备测试视频

如果没有视频文件，生成一个测试视频：

```bash
# 生成一个 10 秒的测试视频
make demo-video

# 或者直接复制你的视频文件
cp /path/to/your/video.mp4 test-videos/
```

### 步骤 3：启动服务

在本地运行（不需要 Docker）：

```bash
make run
```

服务启动后，访问：http://localhost:8080

### 步骤 4：添加转码任务

1. 打开浏览器访问 http://localhost:8080
2. 在表单中输入：
   - 输入文件名：`test-video.mp4`（或你的视频文件名）
   - 可选：码率（如 `1M`）
3. 点击"添加任务"

### 步骤 5：查看结果

查看输出文件夹：

```bash
ls -lh test-output/
```

## 方式二：使用 Docker

### 步骤 1：配置环境变量

```bash
# 复制配置示例
cp .env.example .env

# 编辑 .env 文件，配置你的实际路径
# INPUT_DIR=/home/user/videos
# OUTPUT_DIR=/home/user/output
nano .env
```

### 步骤 2：构建并运行 Docker 容器

```bash
make docker-run
```

### 步骤 3：访问 Web 界面

打开浏览器：http://localhost:8080

### 步骤 4：测试 API

```bash
# 添加转码任务
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"input_file": "a.mp4", "bitrate": "1M"}'

# 查看任务列表
curl http://localhost:8080/api/tasks

# 查看系统统计
curl http://localhost:8080/api/stats
```

## 方式三：快速测试脚本

如果你已经有一个视频文件，使用快速测试脚本：

```bash
# 使用视频文件进行测试
tools/quick-test.sh /path/to/your/video.mp4

# 脚本会自动：
# 1. 检查服务和环境
# 2. 发送转码任务
# 3. 显示任务信息和进度查询命令
```

## 常用命令

```bash
# 开发常用命令
make build          # 编译
make run            # 本地运行
make clean          # 清理

# Docker 相关
make docker-build   # 构建 Docker 镜像
make docker-run     # 运行 Docker 容器
make docker-logs    # 查看日志
make docker-stop    # 停止容器

# 测试
make demo-setup     # 准备测试环境（创建目录）
make demo-video     # 生成测试视频
make demo           # 完整演示流程

# 工具
tools/check-docker.sh           # 检查 Docker 环境
tools/quick-test.sh video.mp4   # 快速测试

# 查看帮助
make help
```

## 验证转码是否成功

### 检查输出文件

```bash
# 查看输出目录
ls -lh test-output/

# 检查视频文件信息（需要 ffmpeg）
ffmpeg -i test-output/your-video_av1.mp4
```

### 检查任务状态

通过 Web 界面或 API：

```bash
# 查看所有任务
curl http://localhost:8080/api/tasks | jq

# 查看特定任务（替换 TASK_ID）
curl http://localhost:8080/api/tasks/your-task-id

# 实时监控（使用 watch）
watch -n 2 'curl -s http://localhost:8080/api/stats | jq'
```

## 常见问题

### Q: 服务启动后无法访问？

A: 检查端口是否被占用：
```bash
netstat -tuln | grep 8080
```

### Q: Docker 权限错误？

A: 将用户添加到 docker 组：
```bash
sudo usermod -aG docker $USER
# 然后重新登录
groups  # 确认包含 docker
```

### Q: 没有视频文件测试？

A: 生成测试视频：
```bash
make demo-video
```

### Q: 转码失败？

A: 查看详细日志：
```bash
# 本地运行
./video-transcoder

# Docker 运行
docker-compose logs -f transcoder
```

### Q: 如何调整并发数？

A: 修改环境变量：
```bash
# 本地
MAX_CONCURRENCY=4 make run

# .env 文件
MAX_CONCURRENCY=4
```

## 参数说明

### 码率设置

- `500k` = 500 kbps（较低质量，文件小）
- `1M` = 1 Mbps（中等质量，平衡）
- `2M` = 2 Mbps（高质量，文件大）

### 并发数建议

| CPU 核心数 | 建议并发数 |
|-----------|-----------|
| 4 核      | 2         |
| 8 核      | 3-4       |
| 16 核以上 | 4-8       |

### 目录结构

本地运行时：
```
video-transcoder/
├── test-videos/      # 输入视频（本地测试）
│   └── video.mp4
├── test-output/      # 输出视频
│   └── video_av1.mp4
└── video-transcoder  # 可执行文件
```

Docker 运行时：
```
宿主机:
/path/to/your/videos/  →  INPUT_DIR
/path/to/output/       →  OUTPUT_DIR
```

## 下一步

- 阅读完整文档： [README.md](README.md)
- 配置生产环境：修改 `.env` 和调整并发数
- 集成到工作流：使用 REST API
- 监控和日志：查看任务状态和错误信息
