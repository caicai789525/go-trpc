package main

import (
	"log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/server"

	_ "trpc.group/trpc-go/trpc-go/http"

	"trpc-go-file-transfer/stub/file_transfer"
)

func main() {
	s := trpc.NewServer()

	// 注册服务
	file_transfer.RegisterFileTransferService(s, &FileTransferServiceImpl{})

	log.Println("文件传输服务启动，监听端口: 8000")

	if err := s.Serve(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
