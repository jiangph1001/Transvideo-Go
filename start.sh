#!/bin/bash
# 本地启动脚本：加载 .env 后运行 video-transcoder
# 用法: ./start.sh
set -e

cd "$(dirname "$0")"

if [ ! -f .env ]; then
  echo "缺少 .env，已从 .env.example 复制，请先修改 INPUT_DIR/OUTPUT_DIR"
  cp .env.example .env
  exit 1
fi

# 导出 .env 中的变量（忽略注释与空行）
set -a
# shellcheck disable=SC1091
source .env
set +a

mkdir -p "$INPUT_DIR" "$OUTPUT_DIR"

if [ ! -x ./bin/video-transcoder ]; then
  echo "二进制不存在，正在编译..."
  PATH="$HOME/.homebrew/bin:$PATH" CGO_ENABLED=1 go build -o bin/video-transcoder main.go
fi

echo "输入目录: $INPUT_DIR"
echo "输出目录: $OUTPUT_DIR"
echo "界面地址: http://localhost:8080"
exec ./bin/video-transcoder
