#!/bin/bash

echo "启动文件传输服务器..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装"
    exit 1
fi

echo "构建项目..."
make build-linux

echo "创建上传目录..."
mkdir -p server1/uploads_server1
mkdir -p server2/uploads_server2

echo "启动服务器1 (端口 8000)..."
cd server1
./file_server1_linux &
SERVER1_PID=$!
cd ..

sleep 3

echo "启动服务器2 (端口 8001)..."
cd server2
./file_server2_linux &
SERVER2_PID=$!
cd ..

echo "服务器启动完成!"
echo "服务器1 PID: $SERVER1_PID, 端口: 8000"
echo "服务器2 PID: $SERVER2_PID, 端口: 8001"
echo "日志输出在各自终端中"

# 等待用户输入停止
echo "按 Ctrl+C 停止服务器"
trap 'echo "停止服务器..."; kill $SERVER1_PID $SERVER2_PID; exit 0' INT

# 等待
wait
