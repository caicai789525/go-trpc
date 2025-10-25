package file_transfer

import (
	"context"

	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/server"
)

// 服务端接口
type FileTransferService interface {
	UploadFile(FileTransfer_UploadFileServer) error
	DownloadFile(context.Context, *FileRequest, FileTransfer_DownloadFileServer) error
	ListFiles(context.Context, *Empty) (*FileList, error)
	DeleteFile(context.Context, *FileRequest) (*OperationResponse, error)
}

func RegisterFileTransferService(s server.Service, svr FileTransferService) {
	s.Register(&FileTransfer_ServiceDesc, svr)
}

var FileTransfer_ServiceDesc = server.ServiceDesc{
	ServiceName: "file_transfer.FileTransfer",
	HandlerType: (*FileTransferService)(nil),
	Methods: []server.Method{
		{
			Name: "ListFiles",
			Func: _FileTransfer_ListFiles_Handler,
		},
		{
			Name: "DeleteFile",
			Func: _FileTransfer_DeleteFile_Handler,
		},
	},
	Streams: []server.StreamDesc{
		{
			StreamName:    "UploadFile",
			Handler:       _FileTransfer_UploadFile_Handler,
			ServerStreams: false,
			ClientStreams: true,
		},
		{
			StreamName:    "DownloadFile",
			Handler:       _FileTransfer_DownloadFile_Handler,
			ServerStreams: true,
			ClientStreams: false,
		},
	},
}

// 处理器函数
func _FileTransfer_UploadFile_Handler(srv interface{}, stream server.Stream) error {
	return srv.(FileTransferService).UploadFile(&fileTransferUploadFileServer{stream})
}

func _FileTransfer_DownloadFile_Handler(srv interface{}, stream server.Stream) error {
	m := &FileRequest{}
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileTransferService).DownloadFile(stream.Context(), m, &fileTransferDownloadFileServer{stream})
}

func _FileTransfer_ListFiles_Handler(srv interface{}, ctx context.Context, f server.Func) (interface{}, error) {
	req := &Empty{}
	if err := f.Decode(req); err != nil {
		return nil, err
	}
	return srv.(FileTransferService).ListFiles(ctx, req)
}

func _FileTransfer_DeleteFile_Handler(srv interface{}, ctx context.Context, f server.Func) (interface{}, error) {
	req := &FileRequest{}
	if err := f.Decode(req); err != nil {
		return nil, err
	}
	return srv.(FileTransferService).DeleteFile(ctx, req)
}

// 流接口
type FileTransfer_UploadFileServer interface {
	SendAndClose(*UploadResponse) error
	Recv() (*FileChunk, error)
	server.Stream
}

type FileTransfer_DownloadFileServer interface {
	Send(*FileChunk) error
	server.Stream
}

type fileTransferUploadFileServer struct {
	server.Stream
}

func (x *fileTransferUploadFileServer) SendAndClose(m *UploadResponse) error {
	return x.Stream.SendMsg(m)
}

func (x *fileTransferUploadFileServer) Recv() (*FileChunk, error) {
	m := &FileChunk{}
	if err := x.Stream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type fileTransferDownloadFileServer struct {
	server.Stream
}

func (x *fileTransferDownloadFileServer) Send(m *FileChunk) error {
	return x.Stream.SendMsg(m)
}

// 客户端
type FileTransferClient interface {
	UploadFile(ctx context.Context, opts ...client.Option) (FileTransfer_UploadFileClient, error)
	DownloadFile(ctx context.Context, req *FileRequest, opts ...client.Option) (FileTransfer_DownloadFileClient, error)
	ListFiles(ctx context.Context, req *Empty, opts ...client.Option) (*FileList, error)
	DeleteFile(ctx context.Context, req *FileRequest, opts ...client.Option) (*OperationResponse, error)
}

type fileTransferClient struct {
	client client.Client
}

func NewFileTransferClientProxy(opts ...client.Option) FileTransferClient {
	return &fileTransferClient{client: client.New(opts...)}
}

type FileTransfer_UploadFileClient interface {
	Send(*FileChunk) error
	CloseAndRecv() (*UploadResponse, error)
	client.ClientStream
}

type FileTransfer_DownloadFileClient interface {
	Recv() (*FileChunk, error)
	client.ClientStream
}

type fileTransferUploadFileClient struct {
	client.ClientStream
}

func (x *fileTransferUploadFileClient) Send(m *FileChunk) error {
	return x.ClientStream.SendMsg(m)
}

func (x *fileTransferUploadFileClient) CloseAndRecv() (*UploadResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := &UploadResponse{}
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type fileTransferDownloadFileClient struct {
	client.ClientStream
}

func (x *fileTransferDownloadFileClient) Recv() (*FileChunk, error) {
	m := &FileChunk{}
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *fileTransferClient) UploadFile(ctx context.Context, opts ...client.Option) (FileTransfer_UploadFileClient, error) {
	stream, err := c.client.NewStream(ctx, &FileTransfer_ServiceDesc.Streams[0], "UploadFile", opts...)
	if err != nil {
		return nil, err
	}
	return &fileTransferUploadFileClient{stream}, nil
}

func (c *fileTransferClient) DownloadFile(ctx context.Context, req *FileRequest, opts ...client.Option) (FileTransfer_DownloadFileClient, error) {
	stream, err := c.client.NewStream(ctx, &FileTransfer_ServiceDesc.Streams[1], "DownloadFile", opts...)
	if err != nil {
		return nil, err
	}
	if err := stream.SendMsg(req); err != nil {
		return nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}
	return &fileTransferDownloadFileClient{stream}, nil
}

func (c *fileTransferClient) ListFiles(ctx context.Context, req *Empty, opts ...client.Option) (*FileList, error) {
	rsp := &FileList{}
	err := c.client.Invoke(ctx, req, rsp, opts...)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (c *fileTransferClient) DeleteFile(ctx context.Context, req *FileRequest, opts ...client.Option) (*OperationResponse, error) {
	rsp := &OperationResponse{}
	err := c.client.Invoke(ctx, req, rsp, opts...)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
