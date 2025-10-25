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

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/errs"
)

const (
	uploadDir   = "./uploads"
	chunkSize   = 64 * 1024         // 64KB
	maxFileSize = 100 * 1024 * 1024 // 100MB
)

type FileTransferServiceImpl struct {
	fileMutex sync.Mutex
}

// 初始化上传目录
func init() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
}

// UploadFile 文件上传
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

		// 第一次接收，创建文件
		if file == nil {
			filename = chunk.Filename
			totalChunks = chunk.TotalChunks

			// 安全检查文件名
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

		// 写入文件块
		n, err := file.Write(chunk.Content)
		if err != nil {
			return errs.New(500, "写入文件失败: "+err.Error())
		}
		receivedBytes += int64(n)

		fmt.Printf("接收文件块 %d/%d, 大小: %d bytes\n",
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

	fmt.Printf("文件上传完成: %s, 总大小: %d bytes\n", filename, receivedBytes)
	return nil
}

// DownloadFile 文件下载
func (s *FileTransferServiceImpl) DownloadFile(req *file_transfer.FileRequest,
	stream file_transfer.FileTransfer_DownloadFileServer) error {

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

		fmt.Printf("发送文件块 %d/%d, 大小: %d bytes\n", i+1, totalChunks, n)
	}

	fmt.Printf("文件下载完成: %s\n", req.Filename)
	return nil
}

// ListFiles 列出所有文件
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

// DeleteFile 删除文件
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

// 安全检查文件名
func isSafeFilename(filename string) bool {
	if filename == "" || strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}
	return true
}
