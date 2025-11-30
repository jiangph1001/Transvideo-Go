.PHONY: build run deps clean test docker-build docker-run docker-stop demo help

# 默认配置
BINARY_NAME=video-transcoder
DOCKER_IMAGE=video-transcoder
GO_VERSION=1.21
BINARY_DIR=bin

# 默认的输入输出目录
DEFAULT_INPUT_DIR=$(PWD)/test-videos
DEFAULT_OUTPUT_DIR=$(PWD)/test-output
DEFAULT_CONCURRENCY=2

help: ## 显示帮助信息
	@echo "Video Transcoder Makefile"
	@echo "========================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## 安装 Go 依赖
	@echo "安装依赖..."
	go mod tidy
	go mod download

build: deps ## 编译 Go 程序
	@echo "编译程序..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) main.go

run: build ## 本地运行程序（需要先准备测试视频）
	@echo "启动视频转码管理系统..."
	@echo "输入目录: $(DEFAULT_INPUT_DIR)"
	@echo "输出目录: $(DEFAULT_OUTPUT_DIR)"
	@echo "并发数: $(DEFAULT_CONCURRENCY)"
	@echo ""
	@echo "正在启动服务，请访问 http://localhost:8080"
	@echo ""
	INPUT_DIR=$(DEFAULT_INPUT_DIR) \
	OUTPUT_DIR=$(DEFAULT_OUTPUT_DIR) \
	MAX_CONCURRENCY=$(DEFAULT_CONCURRENCY) \
	./$(BINARY_DIR)/$(BINARY_NAME)

run-custom: build ## 使用自定义配置运行（需要设置 INPUT_DIR 和 OUTPUT_DIR）
	@echo "使用自定义配置启动..."
	@echo "输入目录: $(INPUT_DIR)"
	@echo "输出目录: $(OUTPUT_DIR)"
	@echo "并发数: $(MAX_CONCURRENCY)"
	@echo ""
	./$(BINARY_DIR)/$(BINARY_NAME)

test: ## 运行测试（如果有测试文件）
	@echo "运行测试..."
	go test -v ./...

clean: ## 清理编译产物
	@echo "清理..."
	rm -rf $(BINARY_DIR)
	go clean

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## 运行 Docker 容器（需要 .env 文件）
	@echo "启动 Docker 容器..."
	@echo "请确保 .env 文件已配置:"
	@echo "  INPUT_DIR=你的视频目录路径"
	@echo "  OUTPUT_DIR=你的输出目录路径"
	@echo ""
	docker-compose up -d

docker-stop: ## 停止 Docker 容器
	@echo "停止 Docker 容器..."
	docker-compose down

docker-logs: ## 查看 Docker 容器日志
	docker-compose logs -f transcoder

demo-setup: ## 准备演示环境（创建测试视频和目录）
	@echo "准备演示环境..."
	@echo "创建测试目录..."
	mkdir -p $(DEFAULT_INPUT_DIR) $(DEFAULT_OUTPUT_DIR)
	@echo ""
	@echo "提示：你需要手动将视频文件复制到 $(DEFAULT_INPUT_DIR)"
	@echo "示例："
	@echo "  cp /path/to/your/video.mp4 $(DEFAULT_INPUT_DIR)/"
	@echo ""
	@echo "或者使用下面的命令生成一个测试视频文件（需要 ffmpeg）："
	@echo "  $(MAKE) demo-video"

demo-video: ## 生成测试视频文件（需要安装 ffmpeg）
	@echo "生成测试视频文件..."
	@echo "检查 ffmpeg..."
	@if ! command -v ffmpeg &> /dev/null; then \
		echo "错误：未找到 ffmpeg，请先安装 ffmpeg"; \
		echo "  Ubuntu/Debian: sudo apt-get install ffmpeg"; \
		echo "  macOS: brew install ffmpeg"; \
		echo "  或访问: https://ffmpeg.org/download.html"; \
		exit 1; \
	fi
	@echo "创建测试目录..."
	mkdir -p $(DEFAULT_INPUT_DIR)
	@echo "生成 10 秒测试视频..."
	ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -pix_fmt yuv420p \
		-c:v libx264 -b:v 2M $(DEFAULT_INPUT_DIR)/test-video.mp4
	@echo "测试视频已生成: $(DEFAULT_INPUT_DIR)/test-video.mp4"
	@echo ""
	@echo "现在你可以运行: make run"

demo: demo-setup ## 完整的演示流程（准备环境并启动）
	@echo ""
	@echo "==================================================="
	@echo "演示环境准备完成！"
	@echo "==================================================="
	@echo ""
	@echo "下一步："
	@echo "1. 将你的视频文件复制到: $(DEFAULT_INPUT_DIR)"
	@echo "2. 运行: make run"
	@echo "3. 访问: http://localhost:8080"
	@echo "4. 在Web界面输入文件名（如: video.mp4）"
	@echo "5. 查看输出: $(DEFAULT_OUTPUT_DIR)"
	@echo ""

check-env: ## 检查环境变量配置
	@echo "当前环境变量配置:"
	@echo "INPUT_DIR=$(INPUT_DIR)"
	@echo "OUTPUT_DIR=$(OUTPUT_DIR)"
	@echo "MAX_CONCURRENCY=$(MAX_CONCURRENCY)"

install-tools: ## 安装常用工具（ffmpeg 等）
	@echo "安装常用工具..."
	@echo "不同系统请使用对应的包管理器:"
	@echo ""
	@echo "Ubuntu/Debian:"
	@echo "  sudo apt-get update"
	@echo "  sudo apt-get install -y ffmpeg docker.io"
	@echo ""
	@echo "macOS:"
	@echo "  brew install ffmpeg"
	@echo ""
	@echo "CentOS/RHEL:"
	@echo "  sudo yum install -y ffmpeg"

fmt: ## 格式化 Go 代码
	@echo "格式化 Go 代码..."
	go fmt ./...

lint: ## 运行 Go 代码检查（需要安装 golangci-lint）
	@echo "运行代码检查..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "警告: golangci-lint 未安装"; \
		echo "安装方法: https://golangci-lint.run/usage/install/"; \
	fi

