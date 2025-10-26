.PHONY: build-windows build-linux build-all clean run-servers

build-windows:
	@echo "构建 Windows 版本..."
	mkdir -p bin
	cd server1 && GOOS=windows GOARCH=amd64 go build -o ../bin/file_server1.exe
	cd server2 && GOOS=windows GOARCH=amd64 go build -o ../bin/file_server2.exe
	cd client && GOOS=windows GOARCH=amd64 go build -o ../bin/file_client.exe
	@echo "构建完成!"

build-linux:
	@echo "构建 Linux 版本..."
	mkdir -p bin
	cd server1 && GOOS=linux GOARCH=amd64 go build -o ../bin/file_server1_linux
	cd server2 && GOOS=linux GOARCH=amd64 go build -o ../bin/file_server2_linux
	cd client && GOOS=linux GOARCH=amd64 go build -o ../bin/file_client_linux
	@echo "构建完成!"

build-all: build-windows build-linux

run-servers:
	@echo "启动服务器..."
	./start_servers.bat

clean:
	rm -rf bin
	@echo "清理完成!"