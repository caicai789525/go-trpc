package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/transport"

	"trpc-go-file-transfer/stub/file_transfer"
)

const (
	serverAddress = "localhost:8000"
	chunkSize     = 64 * 1024 // 64KB
)

type FileTransferClient struct {
	client file_transfer.FileTransferClient
}

func NewFileTransferClient() *FileTransferClient {
	opts := []client.Option{
		client.WithTarget("ip://" + serverAddress),
		client.WithTransport(transport.NewClientTransport("tcp")),
		client.WithTimeout(time.Minute * 10),
	}

	return &FileTransferClient{
		client: file_transfer.NewFileTransferClientProxy(opts...),
	}
}

// UploadFile 上传文件
func (c *FileTransferClient) UploadFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileSize := fileInfo.Size()
	filename := filepath.Base(filePath)
	totalChunks := (fileSize + chunkSize - 1) / chunkSize

	stream, err := c.client.UploadFile(context.Background())
	if err != nil {
		return fmt.Errorf("创建上传流失败: %v", err)
	}

	buffer := make([]byte, chunkSize)
	for i := int64(0); i < totalChunks; i++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件失败: %v", err)
		}

		if n == 0 {
			break
		}

		chunk := &file_transfer.FileChunk{
			Filename:    filename,
			Content:     buffer[:n],
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			FileSize:    fileSize,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("发送文件块失败: %v", err)
		}

		fmt.Printf("上传进度: %d/%d (%.2f%%)\n",
			i+1, totalChunks, float64(i+1)/float64(totalChunks)*100)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("接收响应失败: %v", err)
	}

	if response.Success {
		fmt.Printf("上传成功! 文件路径: %s, 文件大小: %d bytes\n",
			response.FilePath, response.FileSize)
	} else {
		fmt.Printf("上传失败: %s\n", response.Message)
	}

	return nil
}

// DownloadFile 下载文件
func (c *FileTransferClient) DownloadFile(filename, savePath string) error {
	req := &file_transfer.FileRequest{Filename: filename}
	stream, err := c.client.DownloadFile(context.Background(), req)
	if err != nil {
		return fmt.Errorf("创建下载流失败: %v", err)
	}

	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	var receivedBytes int64
	var totalChunks int64

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("接收文件块失败: %v", err)
		}

		n, err := file.Write(chunk.Content)
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		receivedBytes += int64(n)
		totalChunks = chunk.TotalChunks

		fmt.Printf("下载进度: %d/%d (%.2f%%)\n",
			chunk.ChunkIndex+1, totalChunks,
			float64(chunk.ChunkIndex+1)/float64(totalChunks)*100)
	}

	fmt.Printf("下载完成! 文件保存到: %s, 总大小: %d bytes\n", savePath, receivedBytes)
	return nil
}

// ListFiles 列出文件
func (c *FileTransferClient) ListFiles() error {
	empty := &file_transfer.Empty{}
	fileList, err := c.client.ListFiles(context.Background(), empty)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %v", err)
	}

	fmt.Println("服务器上的文件列表:")
	if len(fileList.Files) == 0 {
		fmt.Println("  暂无文件")
		return nil
	}

	for i, filename := range fileList.Files {
		fmt.Printf("  %d. %s\n", i+1, filename)
	}

	return nil
}

// DeleteFile 删除文件
func (c *FileTransferClient) DeleteFile(filename string) error {
	req := &file_transfer.FileRequest{Filename: filename}
	response, err := c.client.DeleteFile(context.Background(), req)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	if response.Success {
		fmt.Printf("文件删除成功: %s\n", filename)
	} else {
		fmt.Printf("文件删除失败: %s\n", response.Message)
	}

	return nil
}
