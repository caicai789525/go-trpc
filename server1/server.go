package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"trpc-go-file-transfer/stub/file_transfer"

	"trpc.group/trpc-go/trpc-go/errs"
)

const (
	uploadDir   = "./uploads_server1"
	chunkSize   = 64 * 1024
	maxFileSize = 100 * 1024 * 1024
)

type FileTransferServiceImpl struct {
	fileMutex sync.Mutex
	transfer  *file_transfer.ServerTransfer
}

// 初始化
func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
}

// 服务器间传输方法
func (s *FileTransferServiceImpl) SyncToServer2(filename string) error {
	localFilePath := filepath.Join(uploadDir, filename)

	if err := s.transfer.SyncFileToServer(localFilePath, "39.96.188.155:8001", filename); err != nil {
		return fmt.Errorf("同步到39.96.188.155失败: %v", err)
	}
	return nil
}

func (s *FileTransferServiceImpl) DownloadFromServer2(filename string) error {
	localSavePath := filepath.Join(uploadDir, "from_server2_"+filename)

	if err := s.transfer.DownloadFileFromServer("39.96.188.155:8001", filename, localSavePath); err != nil {
		return fmt.Errorf("39.96.188.155下载失败: %v", err)
	}
	return nil
}

// 文件传输方法
func (s *FileTransferServiceImpl) UploadFile(stream file_transfer.FileTransfer_UploadFileServer) error {
	var file *os.File
	var receivedBytes int64
	var filename string
	var totalChunks int64

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if file == nil {
			filename = chunk.Filename
			totalChunks = chunk.TotalChunks

			if !isSafeFilename(filename) {
				return errs.New(400, "文件名不安全")
			}

			filePath := filepath.Join(uploadDir, filename)
			file, err = os.Create(filePath)
			if err != nil {
				return errs.New(500, "创建文件失败: "+err.Error())
			}
			defer file.Close()
		}

		n, err := file.Write(chunk.Content)
		if err != nil {
			return errs.New(500, "写入文件失败: "+err.Error())
		}
		receivedBytes += int64(n)

		fmt.Printf("服务器1接收文件块 %d/%d, 大小: %d bytes\n",
			chunk.ChunkIndex+1, totalChunks, len(chunk.Content))
	}

	if err := stream.SendAndClose(&file_transfer.UploadResponse{
		Success:  true,
		Message:  "文件上传成功",
		FilePath: filepath.Join(uploadDir, filename),
		FileSize: receivedBytes,
	}); err != nil {
		return err
	}

	fmt.Printf("服务器1文件上传完成: %s, 总大小: %d bytes\n", filename, receivedBytes)
	return nil
}

func (s *FileTransferServiceImpl) DownloadFile(ctx context.Context, req *file_transfer.FileRequest, stream file_transfer.FileTransfer_DownloadFileServer) error {
	if !isSafeFilename(req.Filename) {
		return errs.New(400, "文件名不安全")
	}

	filePath := filepath.Join(uploadDir, req.Filename)
	file, err := os.Open(filePath)
	if err != nil {
		return errs.New(404, "文件不存在")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return errs.New(500, "获取文件信息失败")
	}

	fileSize := fileInfo.Size()
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	buffer := make([]byte, chunkSize)

	for i := int64(0); i < totalChunks; i++ {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return errs.New(500, "读取文件失败")
		}

		if n == 0 {
			break
		}

		chunk := &file_transfer.FileChunk{
			Filename:    req.Filename,
			Content:     buffer[:n],
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			FileSize:    fileSize,
		}

		if err := stream.Send(chunk); err != nil {
			return err
		}

		fmt.Printf("服务器1发送文件块 %d/%d, 大小: %d bytes\n", i+1, totalChunks, n)
	}

	fmt.Printf("服务器1文件下载完成: %s\n", req.Filename)
	return nil
}

func (s *FileTransferServiceImpl) ListFiles(ctx context.Context, empty *file_transfer.Empty) (*file_transfer.FileList, error) {
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return nil, errs.New(500, "读取目录失败")
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return &file_transfer.FileList{Files: files}, nil
}

func (s *FileTransferServiceImpl) DeleteFile(ctx context.Context, req *file_transfer.FileRequest) (*file_transfer.OperationResponse, error) {
	s.fileMutex.Lock()
	defer s.fileMutex.Unlock()

	if !isSafeFilename(req.Filename) {
		return nil, errs.New(400, "文件名不安全")
	}

	filePath := filepath.Join(uploadDir, req.Filename)
	if err := os.Remove(filePath); err != nil {
		return &file_transfer.OperationResponse{
			Success: false,
			Message: "删除文件失败: " + err.Error(),
		}, nil
	}

	return &file_transfer.OperationResponse{
		Success: true,
		Message: "文件删除成功",
	}, nil
}

func isSafeFilename(filename string) bool {
	if filename == "" || strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}
	return true
}
