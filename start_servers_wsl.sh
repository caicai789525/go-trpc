#!/bin/bash

echo "WSL2 文件传输服务器启动脚本"

# 获取 WSL2 IP 地址
WSL_IP=$(hostname -I | awk '{print $1}')
echo "WSL2 IP 地址: $WSL_IP"

# 构建项目
echo "构建 Linux 版本..."
make build-linux

# 创建上传目录
echo "创建上传目录..."
mkdir -p server1/uploads_server1
mkdir -p server2/uploads_server2

echo "启动服务器1 (端口 8000)..."
cd server1
./file_server1_linux &
SERVER1_PID=$!
cd ..

sleep 2

echo "启动服务器2 (端口 8001)..."
cd server2
./file_server2_linux &
SERVER2_PID=$!
cd ..

echo "=== 服务器启动完成 ==="
echo "服务器1: http://localhost:8000 (PID: $SERVER1_PID)"
echo "服务器2: http://localhost:8001 (PID: $SERVER2_PID)"
echo "WSL2 IP: http://$WSL_IP:8000"
echo "=============================="

# 保存 PID 到文件，便于后续管理
echo $SERVER1_PID > /tmp/file_server1.pid
echo $SERVER2_PID > /tmp/file_server2.pid

echo "按 Ctrl+C 停止服务器"

# 优雅停止
trap 'echo "停止服务器..."; kill $SERVER1_PID $SERVER2_PID; rm -f /tmp/file_server*.pid; exit 0' INT

wait