@echo off
echo 启动文件传输服务器...

echo 启动服务器1 (端口 8000)...
start cmd /k "cd server1 && go run main.go"

timeout /t 3

echo 启动服务器2 (端口 8001)...
start cmd /k "cd server2 && go run main.go"

echo 服务器启动完成!
echo 服务器1: localhost:8000
echo 服务器2: localhost:8001
echo.
echo 按任意键启动客户端...
pause
start cmd /k "cd client && go run main.go"
