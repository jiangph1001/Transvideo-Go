#!/bin/bash

# 检查 Docker 环境脚本

set -e

echo "=========================================="
echo "检查 Docker 环境"
echo "=========================================="
echo ""

echo "1. 检查 Docker 是否安装..."
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安装: $(docker --version)"
else
    echo "❌ Docker 未安装"
    echo "请访问 https://docs.docker.com/get-docker/ 安装 Docker"
    exit 1
fi
echo ""

echo "2. 检查 Docker 服务是否运行..."
if docker info &> /dev/null; then
    echo "✅ Docker 服务正在运行"
else
    echo "❌ Docker 服务未运行"
    echo "请启动 Docker 服务:"
    echo "  Linux: sudo systemctl start docker"
    echo "  macOS: 启动 Docker Desktop"
    echo "  Windows: 启动 Docker Desktop"
    exit 1
fi
echo ""

echo "3. 检查 Docker Socket 权限..."
if [ -w /var/run/docker.sock ]; then
    echo "✅ 当前用户有权限访问 Docker Socket"
else
    echo "⚠️  当前用户可能没有权限访问 Docker Socket"
    echo "解决方案:"
    echo "  1. 将当前用户添加到 docker 组:"
    echo "     sudo usermod -aG docker $USER"
    echo "     然后重新登录 (newgrp docker)"
    echo "  2. 或者使用 sudo 运行 (不推荐)"
    echo "  3. 或者在 docker-compose.yaml 中配置 user: \"root:root\""
fi
echo ""

echo "4. 检查 ffmpeg 镜像..."
if docker image inspect jrottenberg/ffmpeg:ubuntu &> /dev/null; then
    echo "✅ ffmpeg 镜像已存在"
else
    echo "⚠️  ffmpeg 镜像不存在，首次运行时会自动下载"
    echo "    镜像大小约 200-300MB，可能需要一些时间"
fi
echo ""

echo "5. 测试 Docker 运行..."
docker run --rm hello-world > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Docker 运行正常"
else
    echo "❌ Docker 运行异常"
    exit 1
fi
echo ""

echo "=========================================="
echo "✅ 环境检查通过！"
echo "=========================================="
echo ""
echo "现在你可以运行:"
echo "  make docker-run  # 使用 Docker 运行"
echo "  或"
echo "  make demo        # 本地演示"
echo ""
