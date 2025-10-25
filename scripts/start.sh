#!/bin/bash

# 启动 server1
echo "启动 server1..."
cd server1
go run main.go &
SERVER1_PID=$!
cd ..

# 启动 server2
echo "启动 server2..."
cd server2
go run main.go &
SERVER2_PID=$!
cd ..

echo "服务器已启动"
echo "Server1 PID: $SERVER1_PID"
echo "Server2 PID: $SERVER2_PID"

# 等待用户输入停止
echo "按任意键停止服务器..."
read -n 1

# 停止服务器
kill $SERVER1_PID
kill $SERVER2_PID

echo "服务器已停止"