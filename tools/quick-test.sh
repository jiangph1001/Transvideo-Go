#!/bin/bash

# 快速测试脚本
# 用于验证转码系统是否正常工作

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# 检查参数
if [ $# -ne 1 ]; then
    echo "用法: $0 <视频文件名（完整路径或相对路径）>"
    echo ""
    echo "示例:"
    echo "  $0 /home/user/videos/test.mp4"
    echo "  $0 ./test-videos/test.mp4"
    echo ""
    echo "提示:"
    echo "  如果没有视频文件，先运行: make demo-video"
    exit 1
fi

VIDEO_FILE="$1"

# 检查文件是否存在
if [ ! -f "$VIDEO_FILE" ]; then
    echo "❌ 文件不存在: $VIDEO_FILE"
    echo ""
    echo "你可以:"
    echo "  1. 使用现有视频文件"
    echo "  2. 生成测试视频: make demo-video"
    exit 1
fi

echo "=========================================="
echo "快速测试视频转码系统"
echo "=========================================="
echo ""

# 获取文件名
FILENAME=$(basename "$VIDEO_FILE")
NAME_WITHOUT_EXT="${FILENAME%.*}"

# 创建测试目录（如果需要）
INPUT_DIR="$PROJECT_DIR/test-videos"
OUTPUT_DIR="$PROJECT_DIR/test-output"

mkdir -p "$INPUT_DIR" "$OUTPUT_DIR"

# 如果文件不在 INPUT_DIR 中，复制过去
if [ "$(dirname "$VIDEO_FILE")" != "$INPUT_DIR" ]; then
    echo "复制视频文件到测试目录..."
    cp "$VIDEO_FILE" "$INPUT_DIR/"
    echo "✅ 已复制: $INPUT_DIR/$FILENAME"
    echo ""
fi

# 检查服务是否正在运行
echo "检查服务状态..."
if curl -s http://localhost:8080 > /dev/null; then
    echo "✅ 服务正在运行"
else
    echo "❌ 服务未运行"
    echo ""
    echo "请先启动服务:"
    echo "  方法1: make run              # 本地运行"
    echo "  方法2: make docker-run       # Docker运行"
    exit 1
fi
echo ""

# 添加转码任务
echo "添加转码任务..."
echo "视频文件: $FILENAME"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/tasks \
    -H "Content-Type: application/json" \
    -d "{\"input_file\": \"$FILENAME\", \"bitrate\": \"1M\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" == "201" ]; then
    TASK_ID=$(echo "$BODY" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "✅ 任务添加成功!"
    echo "任务ID: $TASK_ID"
    echo ""
    echo "查看状态:"
    echo "  浏览器: http://localhost:8080"
    echo "  API: curl http://localhost:8080/api/tasks/$TASK_ID"
    echo ""
    echo "输出文件将在: $OUTPUT_DIR/${NAME_WITHOUT_EXT}_av1.mp4"
    echo ""
    echo "等待转码完成，你可以运行:"
    echo "  watch -n 2 'curl -s http://localhost:8080/api/tasks/$TASK_ID | grep \"status\"'"
    exit 0
else
    echo "❌ 任务添加失败"
    echo "HTTP 状态码: $HTTP_CODE"
    echo "错误信息: $BODY"
    exit 1
fi
