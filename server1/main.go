package main

import (
	"log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/server"

	"trpc-go-file-transfer/stub/file_transfer"
)

func main() {
	s := trpc.NewServer()

	serviceImpl := &FileTransferServiceImpl{
		transfer: &file_transfer.ServerTransfer{},
	}

	file_transfer.RegisterFileTransferService(s, serviceImpl)

	log.Println("文件传输服务器1启动，监听端口: 8000")

	if err := s.Serve(); err != nil {
		log.Fatalf("服务器1启动失败: %v", err)
	}
}
