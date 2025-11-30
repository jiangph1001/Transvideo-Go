#!/bin/bash

# 快速本地运行脚本（无需 make）
# 适用于快速测试，不需要安装 make 的环境

echo "视频转码管理系统 - 本地运行"
echo "=========================="
echo ""

# 创建测试目录
mkdir -p test-videos test-output

# 检查 ffmpeg 是否存在，如果存在则生成测试视频
if command -v ffmpeg &> /dev/null; then
    if [ ! -f "test-videos/test-video.mp4" ]; then
        echo "生成测试视频..."
        ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -pix_fmt yuv420p \
            -c:v libx264 -b:v 2M test-videos/test-video.mp4 2>&1 | grep -v "^Stream mapping:" | grep -v "^Press \"[q]\"" | grep -v "^  lib"
        echo "✅ 测试视频已生成: test-videos/test-video.mp4"
        echo ""
    fi
else
    echo "⚠️  未安装 ffmpeg，无法生成测试视频"
    echo "   请手动复制视频文件到 test-videos/ 目录"
    echo ""
fi

# 安装 Go 依赖
echo "安装 Go 依赖..."
go mod tidy

# 编译
echo "编译程序..."
go build -o video-transcoder main.go

# 运行
echo ""
echo "启动服务..."
echo "访问 http://localhost:8080"
echo "配置:"
echo "  输入目录: $(pwd)/test-videos"
echo "  输出目录: $(pwd)/test-output"
echo "  并发数: 2"
echo ""

INPUT_DIR=$(pwd)/test-videos \
OUTPUT_DIR=$(pwd)/test-output \
MAX_CONCURRENCY=2 \
./video-transcoder
