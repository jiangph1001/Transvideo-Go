# 项目结构说明

```
video-transcoder/
│
├── main.go                      # 程序入口点
│
├── config/
│   └── config.go                # 配置管理（加载环境变量）
│
├── models/
│   └── task.go                  # 数据模型（Task结构体、状态枚举）
│
├── service/
│   ├── task_service.go          # 任务服务（队列管理、工作池）
│   └── docker_executor.go       # Docker执行器（动态挂载）
│
├── handler/
│   └── task.go                  # HTTP处理器（REST API）
│
├── static/
│   ├── index.html               # 前端页面
│   └── app.js                   # 前端逻辑
│
├── tools/                       # 工具脚本
│   ├── check-docker.sh          # Docker环境检查脚本
│   └── quick-test.sh            # 快速测试脚本
│
├── .env.example                 # 环境变量示例文件
├── .gitignore                   # Git忽略文件
├── Dockerfile                   # Go应用Docker镜像
├── docker-compose.yaml          # Docker编排配置
├── go.mod                       # Go模块依赖
├── Makefile                     # 构建和运行脚本
├── PROJECT_STRUCTURE.md         # 本文件
├── QUICKSTART.md                # 快速开始指南
└── README.md                    # 完整文档
```

## 文件和目录用途详解

### 配置文件

- `.env.example` - 环境变量配置示例，用户需要复制为 `.env` 并修改
- `docker-compose.yaml` - Docker Compose 编排配置，只挂载 docker.sock
- `go.mod` - Go 模块依赖管理
- `Makefile` - 构建、运行、测试的自动化脚本
- `.gitignore` - Git 版本控制忽略列表

### 源代码

#### main.go

程序入口点，负责：
- 加载配置（调用 config.Load()）
- 初始化任务服务（service.NewTaskService()）
- 设置 HTTP 路由（gorilla/mux）
- 启动 HTTP 服务器
- 优雅关闭（处理 SIGINT/SIGTERM 信号）

#### config/config.go

配置管理模块：
- 从环境变量读取配置（INPUT_DIR, OUTPUT_DIR, MAX_CONCURRENCY, DOCKER_HOST）
- 转换为绝对路径
- 返回 Config 结构体

#### models/task.go

数据模型定义：
- Task 结构体：包含任务的所有信息（ID、文件名、状态、时间等）
- TaskStatus 类型：任务状态枚举（pending, running, completed, failed）
- CreateTaskRequest：创建任务的请求结构
- Config：配置结构（可选）

#### service/task_service.go

任务服务核心逻辑：
- 任务队列管理（使用 channel）
- 工作池（goroutine 池）
- 并发控制（可配置最大并发数）
- 任务状态管理
- 任务持久化（在内存 map 中）

#### service/docker_executor.go

Docker 执行器，核心特点：
- **动态挂载**：在启动 ffmpeg 容器时挂载宿主机的目录
- 调用 Docker API（通过命令行）
- 执行转码任务
- 传递 FFmpeg 参数（AV1 编码）

#### handler/task.go

HTTP 请求处理器：
- CREATE /api/tasks - 创建转码任务
- GET /api/tasks - 获取所有任务
- GET /api/tasks/{id} - 获取单个任务
- GET /api/stats - 获取统计信息
- GET / - 提供前端页面

### 前端

#### static/index.html

Web 管理界面：
- 系统状态统计面板
- 添加任务表单
- 任务列表表格
- 自动刷新（5秒）
- 响应式设计

#### static/app.js

前端逻辑：
- API 调用函数
- 表单提交处理
- 任务列表渲染
- 状态标签样式
- 消息提示

### 工具脚本

#### tools/check-docker.sh

Docker 环境检查工具：
- 检查 Docker 是否安装
- 检查 Docker 服务是否运行
- 检查 Docker Socket 权限
- 检查 ffmpeg 镜像是否存在
- 提供解决方案

#### tools/quick-test.sh

快速测试脚本：
- 验证视频文件存在
- 检查服务是否运行
- 发送转码任务请求
- 显示任务信息和后续操作提示

### 构建和部署

#### Dockerfile

Go 应用容器镜像：
- 多阶段构建（builder stage 和 runtime stage）
- 基于 alpine 镜像（体积小）
- 安装 ca-certificates 和 docker-cli
- 创建非 root 用户（transcoder:transcoder）
- 暴露 8080 端口

#### docker-compose.yaml

Docker Compose 编排配置：
- 只挂载 `/var/run/docker.sock`（不挂载视频目录）
- 通过 `env_file` 加载 `.env` 配置
- 配置用户权限（DOCKER_USER, DOCKER_GROUP）
- 可选特权模式
- 远程 Docker 配置示例（注释）

### 文档

#### README.md

完整文档，包含：
- 功能特性
- 系统架构
- 转码参数说明
- 详细使用指南（Web界面和 REST API）
- 配置说明（并发数、远程 Docker）
- 故障排除
- 安全注意事项
- 开发指南
- 性能建议

#### QUICKSTART.md

快速开始指南：
- 5 分钟内快速体验
- 三种启动方式
- 常见问题解答
- 参数说明
- 验证方法

#### PROJECT_STRUCTURE.md

本文件，项目结构说明

## 数据流

### 添加转码任务

```
Web UI / API
    ↓
handler/task.go (HTTP Handler)
    ↓
service/task_service.go (CreateTask)
    ↓
验证文件名格式
验证文件存在
生成任务ID
保存到 tasks map
    ↓
添加到 queue channel
    ↓
返回任务信息
```

### 执行任务

```
Worker goroutine
    ↓
从 queue channel 接收任务
    ↓
service/docker_executor.go (RunTranscodeTask)
    ↓
启动 Docker 容器
    - 挂载 INPUT_DIR:/input
    - 挂载 OUTPUT_DIR:/output
    ↓
执行 FFmpeg 转码
    - 输入: /input/video.mp4
    - 输出: /output/video_av1.mp4
    ↓
更新任务状态
```

### 查询任务

```
Web UI / API
    ↓
handler/task.go
    ↓
service/task_service.go (GetTask / GetAllTasks)
    ↓
返回 tasks map
```

## 配置流程

### 环境变量配置

```bash
# .env 文件
INPUT_DIR=/home/user/videos
OUTPUT_DIR=/home/user/output
MAX_CONCURRENCY=2
DOCKER_HOST=                # 可选：远程 Docker

# 加载流程
1. docker-compose.yaml (env_file: .env)
   ↓
2. 容器环境变量
   ↓
3. config.Load() (读取 os.Getenv)
   ↓
4. service.NewTaskService(cfg)
   ↓
5. docker_executor.RunTranscodeTask()
   ↓
6. 挂载到 ffmpeg 容器
```

## 核心技术设计

### 动态挂载（核心创新）

传统方式：
- 管理容器挂载视频目录 → 启动 ffmpeg 容器 → 共享挂载

本系统：
- 管理容器不挂载视频目录
- 启动 ffmpeg 容器时动态挂载 `-v INPUT_DIR:/input`
- 优点：
  - 更高的安全性
  - 管理容器无文件系统访问权限
  - 配置灵活（只需修改 .env）
  - 资源隔离

### 任务队列

```go
// 使用 Go channel 实现队列
queue := make(chan *models.Task, 1000) // 缓冲区大小 1000

// 工作池模式
for i := 0; i < maxConcurrency; i++ {
    go worker(i)  // 启动多个 worker goroutine
}

// Worker 从队列消费
for {
    select {
    case task := <-queue:
        // 处理任务
    }
}
```

### 并发安全

- `sync.RWMutex` 保护 tasks map
- Channel 保证队列并发安全
- Worker 模式限制并发数

## 部署模式

### 模式 1：本地开发

```bash
make demo              # 准备测试环境
make run               # 本地运行（无需 Docker）
```

### 模式 2：Docker 生产

```bash
cp .env.example .env     # 配置
nano .env               # 修改路径
make docker-run         # 启动服务
```

### 模式 3：远程 Docker

```bash
# .env
DOCKER_HOST=tcp://remote-host:2375
INPUT_DIR=/mnt/nas/videos
OUTPUT_DIR=/mnt/nas/output
```

## 扩展建议

### 添加数据库持久化

- 使用 SQLite/PostgreSQL/MySQL
- 将 tasks map 替换为数据库
- 启动时从数据库加载未完成任务

### 添加监控

- Prometheus metrics
- Grafana 仪表盘
- 任务失败告警

### 添加高级特性

- 任务优先级队列
- 批量任务提交（CSV/JSON）
- 转码参数自定义（前端）
- 进度条（解析 FFmpeg 输出）
- 通知系统（Email/Webhook）
