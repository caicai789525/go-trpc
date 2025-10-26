package main

import (
	"golang.org/x/net/context"
	"log"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/server"
)

func main() {
	s := trpc.NewServer()

	// 直接注册服务实现
	serviceImpl := &FileTransferServiceImpl{}

	// 使用简化的注册方式
	serviceDesc := server.ServiceDesc{
		ServiceName: "file_transfer.FileTransfer",
		HandlerType: (*FileTransferService)(nil),
		Methods: []server.Method{
			{
				Name: "ListFiles",
				Func: func(srv interface{}, ctx context.Context, f interface{}) (interface{}, error) {
					// 简化处理，直接调用服务方法
					return srv.(FileTransferService).ListFiles(ctx, &Empty{})
				},
			},
			{
				Name: "DeleteFile",
				Func: func(srv interface{}, ctx context.Context, f interface{}) (interface{}, error) {
					// 简化处理，直接调用服务方法
					req := &FileRequest{}
					return srv.(FileTransferService).DeleteFile(ctx, req)
				},
			},
		},
		Streams: []server.StreamDesc{
			{
				StreamName: "UploadFile",
				Handler: func(srv interface{}, stream server.Stream) error {
					return srv.(FileTransferService).UploadFile(stream)
				},
				ServerStreams: false,
				ClientStreams: true,
			},
			{
				StreamName: "DownloadFile",
				Handler: func(srv interface{}, stream server.Stream) error {
					req := &FileRequest{}
					if err := stream.RecvMsg(req); err != nil {
						return err
					}
					return srv.(FileTransferService).DownloadFile(stream.Context(), req, stream)
				},
				ServerStreams: true,
				ClientStreams: false,
			},
		},
	}

	s.Register(&serviceDesc, serviceImpl)

	log.Println("文件传输服务器2启动，监听端口: 8001")

	if err := s.Serve(); err != nil {
		log.Fatalf("服务器2启动失败: %v", err)
	}
}
