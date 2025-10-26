package file_transfer

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
)

// ServerTransfer 服务器间文件传输
type ServerTransfer struct {
}

// NewServerTransferClient 创建到其他服务器的客户端
func NewServerTransferClient(target string) FileTransferClient {
	opts := []client.Option{
		client.WithTarget("ip://" + target),
		client.WithTransport(transport.NewClientTransport("tcp")),
		client.WithTimeout(time.Minute * 10),
	}
	return NewFileTransferClientProxy(opts...)
}

// SyncFileToServer 同步文件到另一个服务器
func (st *ServerTransfer) SyncFileToServer(localFilePath, remoteServerAddr, remoteFileName string) error {
	if remoteFileName == "" {
		remoteFileName = filepath.Base(localFilePath)
	}

	// 创建到远程服务器的客户端
	remoteClient := NewServerTransferClient(remoteServerAddr)

	// 上传文件到远程服务器
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileSize := fileInfo.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize

	stream, err := remoteClient.UploadFile(context.Background())
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

		chunk := &FileChunk{
			Filename:    remoteFileName,
			Content:     buffer[:n],
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			FileSize:    fileSize,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("发送文件块失败: %v", err)
		}

		log.Printf("同步文件到服务器 %s: %d/%d (%.2f%%)",
			remoteServerAddr, i+1, totalChunks, float64(i+1)/float64(totalChunks)*100)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("接收响应失败: %v", err)
	}

	if response.Success {
		log.Printf("文件同步成功! 远程服务器: %s, 文件: %s, 大小: %d bytes",
			remoteServerAddr, remoteFileName, response.FileSize)
	} else {
		return fmt.Errorf("文件同步失败: %s", response.Message)
	}

	return nil
}

// DownloadFileFromServer 从另一个服务器下载文件
func (st *ServerTransfer) DownloadFileFromServer(remoteServerAddr, remoteFileName, localSavePath string) error {
	// 创建到远程服务器的客户端
	remoteClient := NewServerTransferClient(remoteServerAddr)

	// 下载文件
	req := &FileRequest{Filename: remoteFileName}
	stream, err := remoteClient.DownloadFile(context.Background(), req)
	if err != nil {
		return fmt.Errorf("创建下载流失败: %v", err)
	}

	file, err := os.Create(localSavePath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %v", err)
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

		log.Printf("从服务器 %s 下载: %d/%d (%.2f%%)",
			remoteServerAddr, chunk.ChunkIndex+1, totalChunks,
			float64(chunk.ChunkIndex+1)/float64(totalChunks)*100)
	}

	log.Printf("从服务器下载完成! 服务器: %s, 文件: %s, 大小: %d bytes",
		remoteServerAddr, remoteFileName, receivedBytes)
	return nil
}

// SyncDirectory 同步整个目录到另一个服务器
func (st *ServerTransfer) SyncDirectory(localDir, remoteServerAddr, remoteDir string) error {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			localFilePath := filepath.Join(localDir, entry.Name())
			remoteFilePath := filepath.Join(remoteDir, entry.Name())

			if err := st.SyncFileToServer(localFilePath, remoteServerAddr, remoteFilePath); err != nil {
				log.Printf("同步文件 %s 失败: %v", entry.Name(), err)
			}
		}
	}

	log.Printf("目录同步完成: %s -> %s:%s", localDir, remoteServerAddr, remoteDir)
	return nil
}
