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
	chunkSize = 64 * 1024 // 64KB
)

type FileTransferClient struct {
	client file_transfer.FileTransferClient
	server string
}

func NewFileTransferClient(serverAddr string) *FileTransferClient {
	opts := []client.Option{
		client.WithTarget("ip://" + serverAddr),
		client.WithTransport(transport.NewClientTransport("tcp")),
		client.WithTimeout(time.Minute * 10),
	}

	return &FileTransferClient{
		client: file_transfer.NewFileTransferClientProxy(opts...),
		server: serverAddr,
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

		fmt.Printf("上传到服务器 %s 进度: %d/%d (%.2f%%)\n",
			c.server, i+1, totalChunks, float64(i+1)/float64(totalChunks)*100)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("接收响应失败: %v", err)
	}

	if response.Success {
		fmt.Printf("上传成功! 服务器: %s, 文件路径: %s, 文件大小: %d bytes\n",
			c.server, response.FilePath, response.FileSize)
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

		fmt.Printf("从服务器 %s 下载进度: %d/%d (%.2f%%)\n",
			c.server, chunk.ChunkIndex+1, totalChunks,
			float64(chunk.ChunkIndex+1)/float64(totalChunks)*100)
	}

	fmt.Printf("下载完成! 服务器: %s, 文件保存到: %s, 总大小: %d bytes\n", c.server, savePath, receivedBytes)
	return nil
}

// ListFiles 列出文件
func (c *FileTransferClient) ListFiles() error {
	empty := &file_transfer.Empty{}
	fileList, err := c.client.ListFiles(context.Background(), empty)
	if err != nil {
		return fmt.Errorf("获取文件列表失败: %v", err)
	}

	fmt.Printf("服务器 %s 上的文件列表:\n", c.server)
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

// SyncBetweenServers 服务器间同步
func (c *FileTransferClient) SyncBetweenServers(filename, targetServer string) error {
	// 先下载文件
	tempPath := "./temp_sync_" + filename
	if err := c.DownloadFile(filename, tempPath); err != nil {
		return fmt.Errorf("下载文件失败: %v", err)
	}
	defer os.Remove(tempPath)

	// 上传到目标服务器
	targetClient := NewFileTransferClient(targetServer)
	if err := targetClient.UploadFile(tempPath); err != nil {
		return fmt.Errorf("上传到目标服务器失败: %v", err)
	}

	fmt.Printf("文件同步成功: %s -> %s\n", c.server, targetServer)
	return nil
}
